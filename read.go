package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var nameLvl = map[int]string{
	0: "EMERG",
	1: "ALERT",
	2: "CRIT",
	3: "ERROR",
	4: "WARNING",
	5: "NOTICE",
	6: "INFO",
	7: "DEBUG",
}

func cErr(ret C.int) string {
	return C.GoString(C.strerror(ret))
}

func msgField(j *C.sd_journal, name string) (string, error) {
	var data unsafe.Pointer
	var length C.size_t

	f := C.CString(name)
	defer C.free(unsafe.Pointer(f))

	if rc := C.sd_journal_get_data(j, f, &data, &length); rc < 0 {
		return "", errors.New("failed to get field: " + cErr(rc))
	}
	fl := C.GoStringN((*C.char)(data), C.int(length))

	return strings.TrimPrefix(fl, name+"="), nil
}

func journalRead(p *service) error {
	var j *C.sd_journal
	rc := C.sd_journal_open(&j, C.SD_JOURNAL_LOCAL_ONLY)
	if rc != 0 {
		return errors.New("error opening journal: " + cErr(rc))
	}
	defer C.sd_journal_close(j)

	// sd_journal_seek_realtime_usec moves the cursor to the entry with the specified timestamp
	rc = C.sd_journal_seek_realtime_usec(j, p.time)
	if rc != 0 {
		return errors.New("error seeking provided until value: " + cErr(rc))
	}

	// match expression to the journal instance _SYSTEMD_UNIT
	sunit := C.CString("_SYSTEMD_UNIT=" + p.unit)
	defer C.free(unsafe.Pointer(sunit))
	rc = C.sd_journal_add_match(j, unsafe.Pointer(sunit), C.strlen(sunit))
	if rc != 0 {
		return errors.New("error setting journal match: " + cErr(rc))
	}

	// inserts a logical OR in the match list
	rc = C.sd_journal_add_disjunction(j)
	if rc < 0 {
		return errors.New("rror set OR match: " + cErr(rc))
	}

	// match expression to the journal instance UNIT
	unit := C.CString("UNIT=" + p.unit)
	defer C.free(unsafe.Pointer(unit))
	rc = C.sd_journal_add_match(j, unsafe.Pointer(unit), C.strlen(unit))
	if rc != 0 {
		return errors.New("error setting journal match: " + cErr(rc))
	}

	for {
		// advances the read pointer into the journal by one entry
		rc = C.sd_journal_next(j)
		if rc < 0 {
			return errors.New("failed to iterate to next entry: " + cErr(rc))
		}
		// EOF
		if rc == 0 {
			break
		}

		// var stamp C.uint64_t
		rc = C.sd_journal_get_realtime_usec(j, &p.time)
		if rc < 0 {
			return errors.New("failed to get realtime timestamp: " + cErr(rc))
		}
		p.time++

		if p.buf.Len() > 3072 {
			continue
		}

		C.sd_journal_restart_data(j)
		msg, err := msgField(j, "PRIORITY")
		if err != nil {
			return err
		}

		lvl, err := strconv.Atoi(msg)
		if err != nil {
			return err
		}
		if lvl > p.level {
			continue
		}
		msg, err = msgField(j, "MESSAGE")
		if err != nil {
			return err
		}
		if len(msg)+p.buf.Len() > 3072 {
			msg = msg[:3072-p.buf.Len()] + "\ncropped message"
		}

		p.buf.WriteString(
			fmt.Sprintf("*%s %s*\n```",
				nameLvl[lvl],
				time.UnixMicro(int64(p.time)).Format("15:04:05")))

		p.buf.WriteString(msg)
		p.buf.WriteString("```")
	}

	return nil
}

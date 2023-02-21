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

	// match expression to the journal instance
	// cmatch := C.CString("_SYSTEMD_UNIT=user@1000.service")
	// defer C.free(unsafe.Pointer(cmatch))
	fmt.Println(p.match)
	rc = C.sd_journal_add_match(j, unsafe.Pointer(p.match), C.strlen(p.match))
	if rc != 0 {
		return errors.New("rror setting journal match: " + cErr(rc))
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

		if p.buf.Len() > 4000 {
			continue
		}

		C.sd_journal_restart_data(j)
		msg, err := msgField(j, "PRIORITY")
		if err != nil {
			return err
		}

		// switch msg {
		// case "0":
		// 	msg = "EMERG "
		// case "1":
		// 	msg = "ALERT "
		// case "2":
		// 	msg = "CRIT "
		// case "3":
		// 	msg = "ERROR "
		// case "4":
		// 	msg = "WRAN "
		// case "5":
		// 	msg = "NOTICE "
		// case "6":
		// 	msg = "INFO "
		// case "7":
		// 	msg = "DEBUG "
		// }

		lvl, err := strconv.Atoi(msg)
		if err != nil {
			return err
		}
		lvl++
		if lvl > p.level {
			continue
		}

		p.buf.WriteString(
			fmt.Sprintf("%s %d | ",
				time.UnixMicro(int64(p.time)).Format("15:04:05"), lvl))

		msg, err = msgField(j, "MESSAGE")
		if err != nil {
			return err
		}
		p.buf.WriteString(msg)
		p.buf.WriteString("\n\n")
		if p.buf.Len() > 4000 {
			p.buf.Truncate(4000)
			p.buf.WriteString("more...")

		}

	}

	return nil
}

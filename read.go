package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
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

func journalRead() error {
	var j *C.sd_journal
	rc := C.sd_journal_open(&j, C.SD_JOURNAL_LOCAL_ONLY)
	if rc != 0 {
		return errors.New("error opening journal: " + cErr(rc))
	}
	defer C.sd_journal_close(j)

	// sd_journal_seek_realtime_usec moves the cursor to the entry with the specified timestamp
	rc = C.sd_journal_seek_realtime_usec(j, C.uint64_t(uint64(time.Now().Add(-time.Hour).UnixMicro())))
	if rc != 0 {
		return errors.New("error seeking provided until value: " + cErr(rc))
	}

	// match expression to the journal instance
	cmatch := C.CString("_SYSTEMD_UNIT=user@1000.service")
	defer C.free(unsafe.Pointer(cmatch))

	rc = C.sd_journal_add_match(j, unsafe.Pointer(cmatch), C.strlen(cmatch))
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

		var stamp C.uint64_t
		rc = C.sd_journal_get_realtime_usec(j, &stamp)
		if rc < 0 {
			return errors.New("failed to get realtime timestamp: " + cErr(rc))
		}

		C.sd_journal_restart_data(j)
		msg, err := msgField(j, "PRIORITY")
		if err != nil {
			return err
		}
		switch msg {
		case "0":
			msg = "EMERG"
		case "1":
			msg = "ALERT"
		case "2":
			msg = "CRIT"
		case "3":
			msg = "ERROR"
		case "4":
			msg = "WRAN"
		case "5":
			msg = "NOTICE"
		case "6":
			msg = "INFO"
		case "7":
			msg = "DEBUG"
		}

		msg, err = msgField(j, "MESSAGE")
		if err != nil {
			return err
		}

	}

	return nil
}

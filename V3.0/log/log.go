//Package log is a custom logger created to log various important information.
package log

import (
	"io"
	"log"
	"os"
)

//Info, Warning, Error, Panic, Fatal were declared.
var (
	Info    *log.Logger //to record information action done by the apps, also it will record the user input mistake.
	Warning *log.Logger //to record information that worth server attention such as fail login attempt or fail action.
	Error   *log.Logger //to record any undesired apps error that are not caused by user.
	Panic   *log.Logger //to record any unexpected panic event.
	Fatal   *log.Logger //to record any fatal event.
)

func init() {

	commonlogFile, err := os.OpenFile("log/commonlog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Fatal.Fatalln("Failed to open common log file:", err)
	}

	errorslogFile, err := os.OpenFile("log/errorslog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Fatal.Fatalln("Failed to open error log file:", err)
	}

	Info = log.New(io.MultiWriter(commonlogFile, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(commonlogFile, os.Stdout), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(errorslogFile, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Panic = log.New(io.MultiWriter(errorslogFile, os.Stderr), "PANIC: ", log.Ldate|log.Ltime|log.Lshortfile)
	Fatal = log.New(io.MultiWriter(errorslogFile, os.Stderr), "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)

}

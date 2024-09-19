package errkit

func Fatal(e error) {
	if e == nil {
		return
	}

	show, file := Output("error.log")
	if file != nil {
		defer func() { _ = file.Close() }()
	}

	show("fatal error %s", e.Error())
}

func Try(e error, protect bool) {
	if e == nil {
		return
	}

	if protect {
		defer func() {
			if cause := recover(); cause == nil {
				return
			} else {
				show, file := Output("error.log")
				show("recover error %v , stack %s", cause, StackTrace(0))
				if file != nil {
					file.Close()
				}
			}
		}()
	}
	panic(e)
}

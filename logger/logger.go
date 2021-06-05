package logger

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

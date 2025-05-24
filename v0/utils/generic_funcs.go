package utils

func GetOpts(opts ...interface{}) interface{} {
	if len(opts) == 0 {
		return nil
	} else if len(opts) == 1 {
		return opts[0]
	} else {
		// TODO: append opts together in one Object, while "overwriting" values. Meaning the later value will overwrite the previous
		return opts[len(opts)-1]
	}
}

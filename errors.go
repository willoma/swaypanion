package swaypanion

type MissingArgumentError struct {
	argument string
}

func (e *MissingArgumentError) Error() string {
	return "Missing argument: " + e.argument
}

func missingArgument(argument string) error {
	return &MissingArgumentError{argument: argument}
}

type WrongArgumentTypeError struct {
	expected string
	value    string
}

func (e *WrongArgumentTypeError) Error() string {
	return `Argument "` + e.value + `" is not ` + e.expected
}

func wrongIntArgument(value string) error {
	return &WrongArgumentTypeError{expected: "an integer", value: value}
}

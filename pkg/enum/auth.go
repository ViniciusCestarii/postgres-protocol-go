package enum_auth

type AuthenticationMethod int

const (
	AuthenticationOk                AuthenticationMethod = 0
	AuthenticationKerberosV5        AuthenticationMethod = 2
	AuthenticationCleartextPassword AuthenticationMethod = 3
	AuthenticationMD5Password       AuthenticationMethod = 5
	AuthenticationGSS               AuthenticationMethod = 7
	AuthenticationGSSContinue       AuthenticationMethod = 8
	AuthenticationSSPI              AuthenticationMethod = 9
	AuthenticationSASL              AuthenticationMethod = 10
	AuthenticationSASLContinue      AuthenticationMethod = 11
	AuthenticationSASLFinal         AuthenticationMethod = 12
)

func (a AuthenticationMethod) String() string {
	switch a {
	case AuthenticationOk:
		return "AuthenticationOk"
	case AuthenticationKerberosV5:
		return "AuthenticationKerberosV5"
	case AuthenticationCleartextPassword:
		return "AuthenticationCleartextPassword"
	case AuthenticationMD5Password:
		return "AuthenticationMD5Password"
	case AuthenticationGSS:
		return "AuthenticationGSS"
	case AuthenticationGSSContinue:
		return "AuthenticationGSSContinue"
	case AuthenticationSSPI:
		return "AuthenticationSSPI"
	case AuthenticationSASL:
		return "AuthenticationSASL"
	case AuthenticationSASLContinue:
		return "AuthenticationSASLContinue"
	case AuthenticationSASLFinal:
		return "AuthenticationSASLFinal"
	default:
		return "Unknown"
	}
}

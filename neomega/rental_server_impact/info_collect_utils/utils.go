package info_collect_utils

import (
	"bufio"
	"fmt"
	"neo-omega-kernel/i18n"
	"neo-omega-kernel/neomega/rental_server_impact/access_helper"
	"neo-omega-kernel/utils/input"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

func LoadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(fmt.Errorf(i18n.T(i18n.S_fail_to_get_user_home_dir), homedir))
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config", "fastbuilder")
	os.MkdirAll(fbconfigdir, 0700)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}

func ReadToken() (string, error) {
	content, err := os.ReadFile(LoadTokenPath())
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func DeleteToken() error {
	return os.Remove(LoadTokenPath())
}

func GetUserInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	return strings.TrimSpace(input), err
}

func GetUserPasswordInput(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	if err != nil {
		return strings.TrimSpace(string(bytePassword)), err
	}
	password := strings.TrimSpace(string(bytePassword))
	if password != "" {
		return password, nil
	} else {
		return GetUserPasswordInput(prompt)
	}
}

func GetRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(i18n.T(i18n.S_please_enter_rental_server_code))
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	code = strings.TrimSpace(strings.TrimRight(code, "\r\n"))
	if code == "" {
		return GetRentalServerCode()
	}
	fmt.Print(i18n.T(i18n.S_please_enter_rental_server_password))
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	password := strings.TrimSpace(string(bytePassword))
	if err != nil {
		return code, password, err
	}
	return code, password, nil
}

func WriteFBToken(token string, tokenPath string) {
	if fp, err := os.Create(tokenPath); err != nil {
		fmt.Printf(i18n.T(i18n.S_fail_to_create_token_file), err)
	} else {
		_, err = fp.WriteString(token)
		if err != nil {
			fmt.Printf(i18n.T(i18n.S_fail_to_write_token), err)
		}
		fp.Close()
	}
}

const (
	AUTH_SERVER_FB_OFFICIAL = "https://api.fastbuilder.pro"
	AUTH_SERVER_LILIYA      = "https://liliya233.uk"
)

var AUTH_SERVER_NAMES = map[string]string{
	AUTH_SERVER_FB_OFFICIAL: i18n.T(i18n.S_auth_server_name_official),
	AUTH_SERVER_LILIYA:      i18n.T(i18n.S_auth_server_name_liliya),
}
var AUTH_SERVER_SELECT_STRINGS = []string{"1. " + i18n.T(i18n.S_auth_server_name_official), "2. " + i18n.T(i18n.S_auth_server_name_liliya)}

func TranslateInputToAuthServer(input string) (authServer string, authServerName string, err error) {
	if strings.Contains(input, "1") {
		return AUTH_SERVER_FB_OFFICIAL, i18n.T(i18n.S_auth_server_name_official), nil
	} else if strings.Contains(input, "2") {
		return AUTH_SERVER_LILIYA, i18n.T(i18n.S_auth_server_name_liliya), nil
	}
	return "", "", fmt.Errorf(i18n.T(i18n.S_invalid_selection_please_select_in_1_to_2))
}

func TranslateAuthServerToAuthServerName(authServer string) (authServerName string) {
	if authServerName, ok := AUTH_SERVER_NAMES[authServer]; ok {
		return authServerName
	} else {
		return fmt.Sprintf(i18n.T(i18n.S_auth_server_name_user_specific), authServer)
	}
}

func ReadUserInfoAndUpdateImpactOptions(impactOptions *access_helper.ImpactOption) (err error) {
	impactOptions.UserName, impactOptions.UserPassword, impactOptions.UserToken,
		impactOptions.ServerCode, impactOptions.ServerPassword,
		impactOptions.AuthServer,
		err =
		ReadUserInfo(
			impactOptions.UserName, impactOptions.UserPassword, impactOptions.UserToken,
			impactOptions.ServerCode, impactOptions.ServerPassword,
			impactOptions.AuthServer,
		)
	// if impactOptions.AuthServer == AUTH_SERVER_FB_OFFICIAL {
	// 	panic(i18n.T(i18n.S_auth_server_version_not_support))
	// }
	return err
}

func ReadUserInfo(userName, userPassword, userToken, serverCode, serverPassword, authServer string) (string, string, string, string, string, string, error) {
	var err error
	authServerName := ""

	flagAuthServerGiven := authServer != ""
	flagUserTokenGiven := userToken != ""
	flagUserNameGiven := userName != ""
	tokenFileToken := ""
	{
		fileUserToken, err := ReadToken()
		if err == nil && fileUserToken != "" {
			tokenFileToken = fileUserToken
		}
	}
	flagHasFileToken := tokenFileToken != ""
	flagNeedInteractivelyInput := false

	// if user token given, and we know the auth server specific -> auth server is given
	if flagUserTokenGiven {
		if !flagAuthServerGiven {
			// if token begins with w9/, then Fastbuilder
			if strings.HasPrefix(userToken, "w9/") {
				authServer = AUTH_SERVER_FB_OFFICIAL
				flagAuthServerGiven = true
				// if token begins with y8/, then Liliya
			} else if strings.HasPrefix(userToken, "y8/") {
				authServer = AUTH_SERVER_LILIYA
				flagAuthServerGiven = true
			}
		}
		if !flagAuthServerGiven {
			fmt.Println(i18n.T(i18n.S_please_input_auth_server_address_or_specific_auth_server))
			authServer = input.GetValidInput()
		}
	} else if flagUserNameGiven {
		// if auth server not given, but userName given, default to Fastbuilder (adapt old codes)
		if !flagAuthServerGiven {
			flagAuthServerGiven = true
			authServer = AUTH_SERVER_FB_OFFICIAL
		}
		authServerName = TranslateAuthServerToAuthServerName(authServer)
		for userPassword == "" {
			userPassword, err = GetUserPasswordInput(fmt.Sprintf(i18n.T(i18n.S_please_input_auth_server_user_password), authServerName))
			if err != nil {
				break
			}
		}
	} else if flagHasFileToken {
		fmt.Print(i18n.T(i18n.S_login_with_current_token))
		if !input.GetInputYN(true) {
			DeleteToken()
			flagNeedInteractivelyInput = true
		} else {
			userToken = tokenFileToken
			if !flagAuthServerGiven {
				// if token begins with w9/, then Fastbuilder
				if strings.HasPrefix(userToken, "w9/") {
					authServer = AUTH_SERVER_FB_OFFICIAL
					flagAuthServerGiven = true
					// if token begins with y8/, then Liliya
				} else if strings.HasPrefix(userToken, "y8/") {
					authServer = AUTH_SERVER_LILIYA
					flagAuthServerGiven = true
				}
			}
			if !flagAuthServerGiven {
				fmt.Println(i18n.T(i18n.S_please_input_auth_server_address_or_specific_auth_server))
				authServer = input.GetValidInput()
			}
		}
	} else {
		flagNeedInteractivelyInput = true
	}

	if flagNeedInteractivelyInput {
		authServerName = TranslateAuthServerToAuthServerName(authServer)
		for authServer == "" {
			fmt.Printf(i18n.T(i18n.S_please_select_auth_server), strings.Join(AUTH_SERVER_SELECT_STRINGS, "\n"))
			selection := input.GetValidInput()
			authServer, authServerName, err = TranslateInputToAuthServer(selection)
			if err != nil {
				pterm.Error.Println(err.Error())
			}
		}
		for userName == "" {
			userName, _ = GetUserInput(fmt.Sprintf(i18n.T((i18n.S_please_input_auth_server_user_name_or_token)), authServerName))
			if strings.Contains(userName, "/") {
				userToken = userName
				userName = ""
				break
			}
		}
		if userToken == "" {
			for userPassword == "" {
				userPassword, err = GetUserPasswordInput(fmt.Sprintf(i18n.T(i18n.S_please_input_auth_server_user_password), authServerName))
				if err == nil {
					break
				}
			}
		}
	}
	// read server code and password
	for serverCode == "" {
		serverCode, serverPassword, err = GetRentalServerCode()
		if err != nil {
			return userName, userPassword, userToken, serverCode, serverPassword, authServer, err
		}
	}
	return userName, userPassword, userToken, serverCode, serverPassword, authServer, nil
}

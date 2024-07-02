package access_helper

import (
	"errors"

	"github.com/OmineDev/neomega-core/i18n"
)

var ErrFBUserCenterLoginFail = errors.New(i18n.T(i18n.S_invalid_auth_server_user_account))
var ErrFBServerConnectionTimeOut = errors.New(i18n.T(i18n.S_auth_server_connection_timeout))
var ErrGetTokenTimeOut = errors.New(i18n.T(i18n.S_get_token_timeout))
var ErrFailToConnectFBServer = errors.New(i18n.T(i18n.S_connection_to_auth_server_timeout))
var ErrRentalServerConnectionTimeOut = errors.New(i18n.T(i18n.S_rental_server_connection_timeout))
var ErrFailToConnectRentalServer = errors.New(i18n.T(i18n.S_fail_to_connect_to_rental_server))
var ErrFBChallengeSolvingTimeout = errors.New(i18n.T(i18n.S_challenge_solving_timeout))
var ErrFBTransferDataTimeout = errors.New(i18n.T(i18n.S_transfer_data_timeout))
var ErrFBTransferDataFail = errors.New(i18n.T(i18n.S_transfer_data_fail))
var ErrFBTransferCheckNumTimeOut = errors.New(i18n.T(i18n.S_transfer_check_num_timeout))
var ErrFBTransferCheckNumFail = errors.New(i18n.T(i18n.S_transfer_check_num_fail))
var ErrBotOpPrivilegeRemoved = errors.New(i18n.T(i18n.S_bot_op_privilege_removed))

func init() {
	i18n.DoAndAddRetranslateHook(func() {
		ErrFBUserCenterLoginFail = errors.New(i18n.T(i18n.S_invalid_auth_server_user_account))
		ErrFBServerConnectionTimeOut = errors.New(i18n.T(i18n.S_auth_server_connection_timeout))
		ErrGetTokenTimeOut = errors.New(i18n.T(i18n.S_get_token_timeout))
		ErrFailToConnectFBServer = errors.New(i18n.T(i18n.S_connection_to_auth_server_timeout))
		ErrRentalServerConnectionTimeOut = errors.New(i18n.T(i18n.S_rental_server_connection_timeout))
		ErrFailToConnectRentalServer = errors.New(i18n.T(i18n.S_fail_to_connect_to_rental_server))
		ErrFBChallengeSolvingTimeout = errors.New(i18n.T(i18n.S_challenge_solving_timeout))
		ErrFBTransferDataTimeout = errors.New(i18n.T(i18n.S_transfer_data_timeout))
		ErrFBTransferDataFail = errors.New(i18n.T(i18n.S_transfer_data_fail))
		ErrFBTransferCheckNumTimeOut = errors.New(i18n.T(i18n.S_transfer_check_num_timeout))
		ErrFBTransferCheckNumFail = errors.New(i18n.T(i18n.S_transfer_check_num_fail))
		ErrBotOpPrivilegeRemoved = errors.New(i18n.T(i18n.S_bot_op_privilege_removed))
	})
}

package i18n

import "fmt"

const (
	S_connecting_to_auth_server                                                                 = "connecting to auth server..."
	S_done_connecting_to_auth_server                                                            = "done connecting to auth server"
	S_connecting_to_mc_server                                                                   = "connecting to mc server (code: %v password: %v) ..."
	S_invalid_auth_server_user_account                                                          = "invalid auth server user account"
	S_rental_server_disconnected                                                                = "rental server disconnected"
	S_auth_server_connection_timeout                                                            = "auth server connection timeout"
	S_get_token_timeout                                                                         = "get token from auth server timeout"
	S_connection_to_auth_server_timeout                                                         = "connect to auth server time out"
	S_rental_server_connection_timeout                                                          = "rental server connection timeout"
	S_fail_to_connect_to_rental_server                                                          = "fail to connect to rental server"
	S_challenge_solving_timeout                                                                 = "challenge solving timeout"
	S_transfer_data_timeout                                                                     = "transfer data timeout"
	S_transfer_data_fail                                                                        = "transfer data fail"
	S_transfer_check_num_timeout                                                                = "transfer check num time out"
	S_transfer_check_num_fail                                                                   = "transfer num fail"
	S_bot_op_privilege_removed                                                                  = "bot privilege removed"
	S_fail_connecting_to_mc_server_retrying                                                     = "fail connecting to mc server, retrying: %v\n"
	S_done_connecting_to_mc_server                                                              = "done connecting to mc server"
	S_coping_with_rental_server_challenges                                                      = "coping with rental server challenges..."
	S_done_coping_with_rental_server_challenges                                                 = "done coping with rental server challenges"
	S_checking_bot_op_permission_and_game_cheat_mode                                            = "checking bot op permission and game cheat mode..."
	S_done_checking_bot_op_permission_and_game_cheat_mode                                       = "done checking bot op permission and game cheat mode"
	S_done_setting_bot_to_creative_mode                                                         = "done setting bot to creative mode"
	S_done_setting_commandblocksenabled_false                                                   = "done setting commandblocksenabled false"
	S_fail_to_get_user_home_dir                                                                 = "fail to get user home dir %v"
	S_please_enter_rental_server_code                                                           = "please enter rental server code: "
	S_please_enter_rental_server_password                                                       = "please enter rental server password: "
	S_fail_to_create_token_file                                                                 = "fail to create token file: %v \n"
	S_fail_to_write_token                                                                       = "fail to write token: %v \n"
	S_auth_server_name_official                                                                 = "FastBuilder Official Auth Server"
	S_auth_server_name_liliya                                                                   = "Liliya233 Auth Server (require individual account)"
	S_invalid_selection_please_select_in_1_to_2                                                 = "invalid input, please input in 1 ~ 2"
	S_auth_server_name_user_specific                                                            = "[auth server: %v]"
	S_please_select_auth_server                                                                 = "please select auth server:\n%v\ninput selection: "
	S_re_login_with_another_account                                                             = "re-login with another account?:"
	S_please_input_auth_server_user_name_or_token                                               = "please input %v username or token: "
	S_please_input_auth_server_user_password                                                    = "please input %v password: "
	S_cannot_find_user_home_dir_save_token_in_current_dir                                       = "cannot find user home dir, token will be saved in current dir"
	S_unknown_err_happened_in_parsing_auth_server_response                                      = "unknown error happened in parsing auth server response: %v"
	S_error_parsing_auth_server_api_response                                                    = "error parsing auth server API response: %v"
	S_cannot_establish_http_connection_with_auth_server_api                                     = "cannot establish http connection with auth server api"
	S_auth_server_is_down_503                                                                   = "auth server is down (http code 503)"
	S_fail_to_fetch_privateSigningKey_from_auth_server                                          = "failed to fetch private signing key from auth server"
	S_fail_to_fetch_keyProve_from_auth_server                                                   = "failed to fetch private keyProve from auth server"
	S_CertSigning_is_disabled_for_errors_above                                                  = "CertSigning is disabled for errors above"
	S_fail_to_transfer_start_type                                                               = "failed to transfer start type: %s"
	S_fail_to_transfer_check_num                                                                = "fail to transfer check num: %s"
	S_fail_to_get_start_type_data                                                               = "failed to get start type data"
	S_fail_to_get_check_num_data                                                                = "failed to get start check num data"
	S_reading_info                                                                              = "reading Info..."
	S_neomega_access_point_starting                                                             = "neOmega access point starting..."
	S_neomega_access_point_ready                                                                = "neOmega access point ready"
	S_fail_to_dispatch_packet_to_remote                                                         = "fail to remote dispatch packets"
	S_no_response_after_a_long_time_bot_is_down                                                 = "no response after %v, bot is down (alive %v)"
	S_bot_operator_privilege_granted                                                            = "bot operator privilege granted"
	S_please_grant_bot_operator_privilege                                                       = "please grant bot operator privilege"
	S_bot_operator_privilege_timeout                                                            = "bot operator privilege granting timed out"
	S_switching_bot_to_creative_mode                                                            = "switching bot to creative mode"
	S_debug_remove_everything_under                                                             = "removing everything under: "
	S_connecting_to_access_point                                                                = "connecting to remote access point..."
	S_connected_to_access_point                                                                 = "connected to remote access point..."
	S_connecting_to_access_point_timeout_rebooting                                              = "connect timeout, rebooting"
	S_autonomous_operator_booting                                                               = "autonomous operator booting..."
	S_autonomous_operator_waiting_for_access_point_connection                                   = "autonomous operator is waiting for access point establish connection to mc server"
	S_enhance_build_is_not_supported                                                            = "enhanced build is not supported"
	S_omega_builder_auth_fail                                                                   = "omega builder auth fail: %v"
	S_omega_builder_api_auth_fail                                                               = "omega builder api auth fail: %v"
	S_programme_logic_error                                                                     = "programme logic error"
	S_not_valid_start_target                                                                    = "not valid start target, below are provided start target:"
	S_please_input_auth_server_address_or_specific_auth_server                                  = "please input auth server address or use -A to specific auth server when programme start"
	S_login_with_current_token                                                                  = "login with current token (y)? or use new account (n)? please input y or n:"
	S_net_position_conflict_only_one_access_point_can_exist                                     = "net position conflict, only on access point should exist on this network"
	S_send_commandfeedback_is_set_to_be                                                         = "send command feedback is set to be %v"
	S_detecting_send_command_feedback_mode                                                      = "detecting send command feedback in rental server"
	S_no_access_point_in_network                                                                = "no access point in network"
	S_device_performance_insufficient_msg_dropped                                               = "device performance insufficient, msg dropped"
	S_starting_post_challenge_actions                                                           = "staring post-challenge actions: %v\n"
	S_omega_launcher_config_file_and_programme_logic_mismatch_or_programme_logic_error_in_stage = "neOmega config file and programme logic mismatch / error exist in programme logic. stage: %v"
	S_executing_login_sequence                                                                  = "executing login sequence..."
	S_updating_tag_in_omega_net                                                                 = "updating tag in omega network"
	S_generating_client_key_pair                                                                = "generating client key pair"
	S_retrieving_client_information_from_auth_server                                            = "retrieving client information from auth server"
	S_establishing_raknet_connection                                                            = "establishing raknet connection"
	S_establishing_byte_frame_connection                                                        = "establishing byte frame connection"
	S_generating_packet_connection                                                              = "generating packet connection"
	S_generating_key_login_request                                                              = "generating key login request"
	S_exchanging_login_data                                                                     = "exchanging login data"
	S_login_accomplished                                                                        = "login accomplished"
	S_sending_additional_information                                                            = "sending additional information"
	S_packing_core                                                                              = "packing core"

	S_unknown_code                        = "unknown code: %v"
	C_Auth_BackendError                   = 5
	C_Auth_FailedToRequestEntry           = 6
	C_Auth_HelperNotCreated               = 7
	C_Auth_InvalidFBVersion               = 8
	C_Auth_InvalidHelperUsername          = 9
	C_Auth_InvalidToken                   = 10
	C_Auth_InvalidUser                    = 11
	C_Auth_ServerNotFound                 = 12
	C_Auth_UnauthorizedRentalServerNumber = 13
	C_Auth_UserCombined                   = 14
	C_Auth_FailedToRequestEntry_TryAgain  = 15
)

var I18nDict_zh_cn map[string]string = map[string]string{
	S_executing_login_sequence:                                 "开始执行登陆序列",
	S_updating_tag_in_omega_net:                                "正在检查并更新本地neOmega网络中的属性信息",
	S_generating_client_key_pair:                               "正在生成客户端密钥对",
	S_retrieving_client_information_from_auth_server:           "正在从验证服务器取得机器人数据",
	S_establishing_raknet_connection:                           "正在建立 raknet 连接",
	S_establishing_byte_frame_connection:                       "正在封装数据帧连接层",
	S_generating_packet_connection:                             "正在封装数据包连接层",
	S_generating_key_login_request:                             "正在生成关键登陆数据",
	S_exchanging_login_data:                                    "正在和 Minecraft 服务器交换登陆信息",
	S_login_accomplished:                                       "登陆序列完成",
	S_sending_additional_information:                           "正在发送附加信息",
	S_packing_core:                                             "正在打包关键数据",
	S_connecting_to_auth_server:                                "正在连接到验证服务器",
	S_done_connecting_to_auth_server:                           "成功与验证服务器建立连接",
	S_connecting_to_mc_server:                                  "正在连接到网易租赁服(服号: %v 密码: %v)...",
	S_invalid_auth_server_user_account:                         "无效的验证服务器账户 -> 请确认没有选错验证服务器，且没有输错账户名",
	S_rental_server_disconnected:                               "与网易租赁服连接已经断开 -> 网易土豆的常见问题，检查你的租赁服状态（等级、是否开启、密码）并重试",
	S_auth_server_connection_timeout:                           "与验证服务器连接超时 -> 检查是否选错验证服务器和你的网络状态，若重试还是不行，可能需要切换网络或者验证服务器挂了",
	S_get_token_timeout:                                        "从验证服务器获得 token 超时 -> 非常规的错误，可能是你的账户/密码错误，或者网络连接已经断开",
	S_connection_to_auth_server_timeout:                        "与验证服务器连接超时 -> 检查是否选错验证服务器和你的网络状态，若重试还是不行，可能需要切换网络或者验证服务器挂了",
	S_rental_server_connection_timeout:                         "与网易租赁服连接超时 -> 网易土豆的常见问题，检查你的租赁服状态（等级、是否开启、密码）并重试, 也可能是你的网络问题",
	S_fail_to_connect_to_rental_server:                         "无法连接到网易租赁服 -> 网易土豆的常见问题，检查你的租赁服状态（等级、是否开启、密码）并重试, 也可能是你的网络问题",
	S_challenge_solving_timeout:                                "无法在指定时限内完成机器人身份的零知识证明 -> 请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_transfer_data_timeout:                                    "验证服务器无法在指定时限内代为完成网易要求的机器人身份零知识证明的[data]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_transfer_data_fail:                                       "验证服务器无法代为完成网易要求的机器人身份零知识证明的[data]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_transfer_check_num_timeout:                               "验证服务器无法在指定时限内代为完成网易要求的机器人身份零知识证明的[check num]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_transfer_check_num_fail:                                  "验证服务器无法代为完成网易要求的机器人身份零知识证明的[check num]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_bot_op_privilege_removed:                                 "机器人的 OP 权限突然被移除 -> 不知道是谁/什么东西突然把机器人 /deop 了，重启应该可以解决此问题",
	S_fail_connecting_to_mc_server_retrying:                    "无法连接到网易租赁服, 正在第 %v 次重新尝试连接到网易租赁服 -> 请查看之前的输出，可能指明了无法连接的实际原因\n",
	S_done_connecting_to_mc_server:                             "成功连接到网易租赁服",
	S_coping_with_rental_server_challenges:                     "正在尝试完成网易要求的零知识机器人身份证明...",
	S_done_coping_with_rental_server_challenges:                "成功完成网易要求的零知识机器人身份证明",
	S_checking_bot_op_permission_and_game_cheat_mode:           "正在检查机器人的权限和租赁服内作弊模式是否打开...",
	S_done_checking_bot_op_permission_and_game_cheat_mode:      "机器人的权限和租赁服内作弊模式检查完成，无问题",
	S_done_setting_bot_to_creative_mode:                        "成功将机器人设置为创造模式",
	S_done_setting_commandblocksenabled_false:                  "成功关闭命令块 (gamerule commandblocksenabled false)",
	S_fail_to_get_user_home_dir:                                "无法获得用户的家(home, ~/)目录:  %v -> 这应该不是大问题",
	S_please_enter_rental_server_code:                          "请输入网易租赁服号: ",
	S_please_enter_rental_server_password:                      "请输入网易租赁服密码(没有则留空): ",
	S_fail_to_create_token_file:                                "无法创建 token 文件 -> 这应该不是大问题\n",
	S_fail_to_write_token:                                      "无法写入 token 文件 -> 这应该不是大问题\n",
	S_auth_server_name_official:                                "FastBuilder 官方验证服务器",
	S_auth_server_name_liliya:                                  "咕咕酱的验证服务器 (需要独立账户)",
	S_invalid_selection_please_select_in_1_to_2:                "无效的输入，请输入 1 ~ 2 间的选择",
	S_auth_server_name_user_specific:                           "[验证服务器: %v]",
	S_please_select_auth_server:                                "请选择验证服务器:\n%v\n请输入编号: ",
	S_re_login_with_another_account:                            "需要使用其它账户(而非现有 token 中的账户)登陆吗? (y/n): ",
	S_please_input_auth_server_user_name_or_token:              "请输入验证服务器[%v]的账户或 token: ",
	S_please_input_auth_server_user_password:                   "请输入验证服务器[%v]的用户密码 (若密码无法登陆，请尝试使用 token 或者使用一次性密码), 密码不会显示: ",
	S_cannot_find_user_home_dir_save_token_in_current_dir:      "无法找到用户家 (home) 目录，token 将被保存在当前目录",
	S_unknown_err_happened_in_parsing_auth_server_response:     "在解析验证服务器响应时发生了未知错误: %v  -> 请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_error_parsing_auth_server_api_response:                   "无法解析验证服务器响应: %v  -> 请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_cannot_establish_http_connection_with_auth_server_api:    "无法与验证服务器建立 http 连接 -> 检查是否选错验证服务器和你的网络状态，若重试还是不行，可能需要切换网络或者验证服务器挂了",
	S_auth_server_is_down_503:                                  "验证服务器已宕机 (http 响应码: 503) -> 请联系验证服务器维护者，你大概是解决不了这个问题的",
	S_fail_to_fetch_privateSigningKey_from_auth_server:         "无法从验证服务器获得 privateSigningKey -> 无所谓，反正你不用搞懂那个是个什么东西 （在给bdx签名时才用的到）",
	S_fail_to_fetch_keyProve_from_auth_server:                  "无法从验证服务器获得 keyProve -> 无所谓，反正你不用搞懂那个是个什么东西 （在给bdx签名时才用的到）",
	S_CertSigning_is_disabled_for_errors_above:                 "由于上述错误，已禁用 CertSigning -> 无所谓，反正你不用搞懂那个是个什么东西 （在给bdx签名时才用的到）",
	S_fail_to_transfer_start_type:                              "无法向验证服务器转发网易要求的[start type]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_fail_to_transfer_check_num:                               "无法向验证服务器转发网易要求的机器人身份零知识证明的[check num]项目 ->  请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	S_fail_to_get_start_type_data:                              "无法从网易服务器获得网易要求的[start type]挑战 -> 这是网易服务器问题导致的，请重启试试，若错误一直出现可能是因为协议发生变化需要开发者适配",
	S_fail_to_get_check_num_data:                               "无法从网易服务器获得网易要求的[check num]挑战 -> 这是网易服务器问题导致的，请重启试试，若错误一直出现可能是因为协议发生变化需要开发者适配",
	S_unknown_code:                                             "未知的验证服务器响应:%v -> 截屏然后发给 neOmega 开发者，让他们适配",
	S_reading_info:                                             "读取登陆配置中...",
	S_neomega_access_point_starting:                            "[neOmega 接入点]: 启动中...",
	S_neomega_access_point_ready:                               "[neOmega 接入点]: 就绪",
	S_fail_to_dispatch_packet_to_remote:                        "[neOmega 接入点] -> [neOmega 访问点] 数据包分发失败",
	S_no_response_after_a_long_time_bot_is_down:                "持续 %v 秒未能从网易租赁服获得数据, 机器人已确认掉线 (机器人在线时间 %vs)",
	S_bot_operator_privilege_granted:                           "机器人已获得操作员权限",
	S_please_grant_bot_operator_privilege:                      "缺少许可，请给予许可",
	S_bot_operator_privilege_timeout:                           "机器人未能在指定时限内获得操作员权限 -> 在机器人再次上线时，请及时给予机器人操作员权限",
	S_switching_bot_to_creative_mode:                           "正在将机器人切换为创造模式",
	S_debug_remove_everything_under:                            "Debug: 正在移除目录下所有 neOmega 文件",
	S_connecting_to_access_point:                               "[neOmega 访问点]: 正在连接到 [neOmega 接入点]",
	S_connected_to_access_point:                                "[neOmega 访问点]: 成功连接到 [neOmega 接入点]",
	S_connecting_to_access_point_timeout_rebooting:             "[neOmega 访问点]: 与 [neOmega 接入点] 的连接超时, 正在重连",
	S_autonomous_operator_booting:                              "自律型 Operator 启动中...",
	S_autonomous_operator_waiting_for_access_point_connection:  "自律型 Operator 正在等待 [neOmega 接入点] 建立与 Minecraft 服务器的连接",
	S_enhance_build_is_not_supported:                           "您使用的程序构建并不包含 neOmega 导入模块 -> 这是因为您没有从正规渠道下载，正规渠道下载的应该包含 neOmega 导入模块",
	S_omega_builder_auth_fail:                                  "neOmega 导入器授权失败: %v -> 您未获得neOmega导入器使用权限，请联系销售人员",
	S_omega_builder_api_auth_fail:                              "neOmega 导入器 API 授权失败: %v -> 您未获得neOmega导入器 API 使用权限，请联系销售人员",
	S_programme_logic_error:                                    "程序设计存在缺陷 -> 请截屏并将错误信息发给 neOmega 开发者并等待此问题修复",
	S_not_valid_start_target:                                   "不是有效的启动目标，以下为可选的的启动目标:",
	S_please_input_auth_server_address_or_specific_auth_server: "请输入验证服务器地址或者使用 -A 参数在程序启动时指定验证服务器",
	S_login_with_current_token:                                 "使用先前保存的 Token 登陆(y)? 还是更换 Token (n)? 请输入 y 或 n:",
	S_net_position_conflict_only_one_access_point_can_exist:    "网络位置冲突, 网络中已经存在 [neOmega 接入点] -> 此问题很可能是程序逻辑错误，请截屏联系开发者",
	S_send_commandfeedback_is_set_to_be:                        "根据启动时信息, omega 将沿用设置: gamerule sendcommandfeedback=%v",
	S_detecting_send_command_feedback_mode:                     "正在检测租赁服内 sendcommandfeedback 模式",
	S_no_access_point_in_network:                               "[neOmega 接入点] 不存在于网络中或未启动 -> 请先启动 [neOmega 接入点]",
	S_device_performance_insufficient_msg_dropped:              "设备性能不足或服务器压力过大, 数据已丢弃 -> 考虑使用更好的设备或降低工作负载",
	S_starting_post_challenge_actions:                          "执行后挑战任务: %v\n",
	S_omega_launcher_config_file_and_programme_logic_mismatch_or_programme_logic_error_in_stage: "neOmega 启动器逻辑存在缺陷或配置文件版本与启动器不符(%v) -> 程序尝试自动排除错误，请重启，若错误反复出现请联系 neOmega 开发者反馈问题",
}

var I18nCodeDict_zh_cn map[int]string = map[int]string{
	C_Auth_BackendError:                   "验证服务器: 后端错误 -> 请重试，若一直出现该错误请联系验证服务器维护者，你自己大概是修不了这个的",
	C_Auth_FailedToRequestEntry:           "验证服务器: 验证服务器无法获得你的网易租赁服入口 -> 请进入游戏检查你的网易租赁服是否能搜索到?是否公开?是否有等级限制和密码? 然后重试",
	C_Auth_HelperNotCreated:               "验证服务器: 你还没有创建机器人 -> 请进入验证服务器对应网站， 根据网站上的指示创建机器人，此程序需要一个机器人进入网易租赁服并执行任务",
	C_Auth_InvalidFBVersion:               "验证服务器: 你使用的程序与验证服务器协议版版本不匹配 -> 请更新此程序, 若更新后此问题仍然出现，请让 neOmega 开发者适配协议",
	C_Auth_InvalidHelperUsername:          "验证服务器: 你的账户的机器人的游戏内名称无效 -> 请进入验证服务器对应网站， 根据网站上的指示重命名机器人，或者重新创建机器人。因为你的机器人名字内可能包含敏感词或者其它不利于进入服务器的因素",
	C_Auth_InvalidToken:                   "验证服务器: 无效的 token -> 注意，token 会随你的密码改变而失效，你应该重新从验证服务器对应网站上下载token，或使用账户名密码登陆，当然，请确保你的验证服务器选择没错",
	C_Auth_InvalidUser:                    "验证服务器: 无效的账户 -> 你的账户名密码是否正确？你的验证服务器选择是否与账户对应？",
	C_Auth_ServerNotFound:                 "验证服务器: 验证服务器无法找到你的网易租赁服 -> 请进入游戏检查你的网易租赁服是否能搜索到? 是否公开? 然后重试",
	C_Auth_UnauthorizedRentalServerNumber: "验证服务器: 此租赁服未在验证服务器中授权 -> 请登陆验证服务器对应网站或联系销售者获得授权",
	C_Auth_UserCombined:                   "验证服务器: 此账户已经被合并到另一个账户，这个账户已经无法使用了，请使用另一个账户登陆 -> 如果你不知道发生了什么，那你应该是被谁骗了，下次千万别再上当了，保护好自己账户",
	C_Auth_FailedToRequestEntry_TryAgain:  "验证服务器: 验证服务器无法获得你的网易租赁服入口 -> 请进入游戏检查你的网易租赁服是否能搜索到?是否公开?是否有等级限制和密码? 然后重试",
}

var I18nDict map[string]string = I18nDict_zh_cn
var I18nCodeDict map[int]string = I18nCodeDict_zh_cn
var hooks []func()

func T(code string) string {
	r, has := I18nDict[code]
	if !has {
		return code
	}
	return r
}

func CT(code int) string {
	r, has := I18nCodeDict[code]
	if !has {
		return fmt.Sprintf(T(S_unknown_code), code)
	}
	return r
}

func DoAndAddRetranslateHook(hook func()) {
	hook()
	hooks = append(hooks, hook)
}

func ChangeLanguage() {
	I18nDict = I18nDict_zh_cn
	I18nCodeDict = I18nCodeDict_zh_cn
	for _, hook := range hooks {
		hook()
	}
}

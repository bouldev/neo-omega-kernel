try:
    # 如果在程序顶层
    from conn import ThreadOmega,ConnectType,AccountOptions
except:
    try:
        # 如果被放到了包里
        from .conn import ThreadOmega,ConnectType,AccountOptions
    except:
        # 我不想污染环境变量，但如果程序运行到这里那肯定是你的错 --240
        import sys 
        import os
        sys.path.append(os.path.dirname(__file__))
        try:
            # 如果在程序顶层
            from conn import ThreadOmega,ConnectType,AccountOptions
        except:
            from .conn import ThreadOmega,ConnectType,AccountOptions
            # 居然这样还不行？ 反正我尽力了，这肯定是你的错

if __name__ == '__main__':
    # 直接在内部启动一个 neOmega, 不需要 fb,omega 也不需要新进程
    # 你可以把它当成一个普通函数
    # 因为是在内部启动的，所以需要账号密码
    # 为什么明明是在本地，还有address? 这个是为了方便远程连接的, 你可以新开一个远程连接 ConnectType.Remote 它会连接到这个进程
    # 这样, python conn.py 就可以远程连接了
    omega=ThreadOmega(
        connect_type=ConnectType.Local,
        address="tcp://localhost:24015",
        accountOption=AccountOptions(
            UserName="2401PT",
            UserPassword="24*******",
            ServerCode="96******"
        )
    )
    omega.wait_disconnect()
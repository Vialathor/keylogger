def main():
    device_id = input("Device id: ")

    while True:
        cmd = input("Enter cmd: Start | Upload | Stop").strip().lower()
        
        if cmd == "start":
            update_cmd(device_id, cmd)
        elif cmd == "upload":
            update_cmd(device_id, cmd)
        elif cmd == "stop":
            update_cmd(device_id, cmd)
        else:
            break


def update_cmd(device_id, cmd):
    pass
    
if __name__ == "__main__":
    main()
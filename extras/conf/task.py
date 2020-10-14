#!/usr/bin/env python3

import sys
from os import write
from sys import argv
from os import system
from uuid import uuid4
from os import listdir
from datetime import datetime
from json import dumps, loads
from shutil import copytree, move
from argparse import ArgumentParser
from os.path import join, exists, isdir

done='done'
ongoing='ongoing'
template='template'


def task_add_fn(name):
    # TODO: directory where this file is
    src = template
    dst = join(ongoing, name)

    if exists(dst):
        raise Exception('task %s exist' % name)

    copytree(src, dst)

    config = join(dst, '.config')
    tags = join(dst, '.tags')

    data = dict()
    data['Task-Name'] = name
    data['Task-Id'] = uuid4().hex

    with open(config, 'w') as fo:
        fo.write('{}\n'.format(dumps(data, indent=4)))

    with open(tags, 'a') as fo:
        fo.write('{}\n'.format(name))

    task_info(name)

def task_info(name):
    src = join(ongoing, name)
    config = join(src, '.config')
    tags = join(src, '.tags')

    with open(config, 'r') as fo:
        data = loads(fo.read())
    print('\n'.join(['{}: {}'.format(k, v) for k, v in data.items()]))

    with open(tags, 'r') as fo:
        lines = list()
        for line in fo.readlines():
            lines.append(line.strip())
    print('Tags: %s' % ', '.join(lines))

def task_info_fn(name):
    if name:
        task_info(name)
        return
    for name in listdir(ongoing):
        if isdir(join(ongoing, name)):
            print('*' * 20)
            task_info(name)

def task_done_fn(name):
    src = join(ongoing, name)
    dst = join(done, name)

    if not exists(src):
        raise Exception("task %s doesn't exist" % name)
    move(src, dst, copy_function=copytree)

def main(args):
    args.fn(args.name)


if __name__ == '__main__':
    parser = ArgumentParser()

    parsers = parser.add_subparsers(dest='action', required=True)

    p = parsers.add_parser('add')
    p.set_defaults(fn=task_add_fn)
    p.add_argument('name')

    p = parsers.add_parser('done')
    p.set_defaults(fn=task_done_fn)
    p.add_argument('name')

    p = parsers.add_parser('info')
    p.set_defaults(fn=task_info_fn)
    p.add_argument('--name')

    args = parser.parse_args()

    try:
        main(args)
    except KeyboardInterrupt:
        sys.exit(1)
    except Exception as err:
        sys.stderr.write('error|%s|caught unhandled exception\n' % (
            datetime.now().strftime('%d-%m-%Y %H:%M:%S')
        ))
        err = str(err)
        sys.stderr.write(err if err.endswith('\n') else err + '\n')
        sys.exit(1)
    sys.exit(0)


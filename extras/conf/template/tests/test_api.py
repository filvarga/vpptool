#!/bin/env python
from __future__ import print_function
import atexit, fnmatch, os, sys
from vpp_papi import VPP 


class BAPI(object):

    json_dirs = [
        'build-root/install-vpp_debug-native/vpp/share/vpp/api/core',
        'build-root/install-vpp_debug-native/vpp/share/vpp/api/plugins'
        ]

    vpp_dir = os.environ['VPP']
   
    def __init__(self):
        self.json_files = list()

        for json_dir in self.json_dirs:
            self.construct(json_dir)

        if not self.json_files:
            sys.stderr.write('error: no json api files found\n')
            sys.exit(1)

    def construct(self, json_dir):
        json_dir = os.path.join(self.vpp_dir, json_dir)
        for _, _, file_names in os.walk(json_dir):
            for file_name in fnmatch.filter(file_names, '*.api.json'):
                self.json_files.append(os.path.join(json_dir, file_name))

    @classmethod
    def connect(cls, connection_name='api-test'):
        obj = cls()
        vpp = VPP(obj.json_files)
        r = vpp.connect(connection_name)

        if r != 0:
            sys.stderr.write('error: connecting to VPP\n')
            sys.exit(r)

        atexit.register(vpp.disconnect)
       
        r = vpp.api.show_version()
        sys.stdout.write('connected to VPP {}\n'.format(r.version.decode().rstrip('\0x00')))
        return vpp


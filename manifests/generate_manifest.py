#!/usr/bin/env python3
import json
import os
from subprocess import check_output

DEFAULT_IMAGE_NAME_PREFIX = 'openshift-'
DEFAULT_IMAGE_REMOTE = 'registry-proxy.engineering.redhat.com'


def brew_build_info(nvr):
    return check_output(
        [
            'brew',
            'call',
            'getBuild',
            nvr,
            '--json-output'
        ]
    )


def brew_latest_build(tag, name):
    return check_output(
        [
            'brew',
            'latest-build',
            '--quiet',
            tag,
            name
        ]
    )


def get_build_info(tag, name):
    latest_build = brew_latest_build(tag, name)
    print('get latest build:', latest_build)
    nvr = latest_build.split()[0]
    build_info = brew_build_info(nvr)
    print('get build info:', build_info)
    return build_info


def get_image_manifest(component_name,
                       build_info,
                       image_name_prefix=DEFAULT_IMAGE_NAME_PREFIX,
                       image_remote=DEFAULT_IMAGE_REMOTE):

    if build_info:
        obj = json.loads(build_info)
        return {
            "image-key":      component_name,
            "image-name":     obj["extra"]["image"]["index"]["pull"][0].split('@')[0].split('/')[-1].replace(image_name_prefix, ''),  # noqa
            "image-version":  obj["extra"]["image"]["index"]["floating_tags"][1],
            "image-tag":      obj["extra"]["image"]["index"]["tags"][0],
            "image-remote":   image_remote,
            "image-digest":   obj["extra"]["image"]["index"]["pull"][0].split('@')[1]  # noqa
        }
    else:
        return {
            "image-key":      component_name,
            "image-name":     'ERROR: build_info is empty',
            "image-version":  'ERROR: build_info is empty',
            "image-tag":      'ERROR: build_info is empty',
            "image-remote":   image_remote,
            "image-digest":   'ERROR: build_info is empty'
        }


def main():
    operands = [
        {
            "component_name": "logging-fluentd",
            "build_info": os.getenv('LOGGING_FLUENTD_BUILD_INFO_JSON')
        },
        {
            "component_name": "logging-curator5",
            "build_info": os.getenv('ELASTICSEARCH_OPERATOR_BUILD_INFO_JSON')
        },
        {
            "component_name": "cluster-logging-operator",
            "build_info": os.getenv('CLUSTER_LOGGING_OPERATOR_BUILD_INFO_JSON')
        },
    ]
    # no extarnal operands for cluster-logging
    external_operands = []
    manifest = []
    for operand in operands:
        print('processing', operand['component_name'])
        manifest.append(get_image_manifest(
            operand['component_name'],
            operand['build_info'])
        )
    for operand in external_operands:
        print('processing', operand['component_name'])
        manifest.append(get_image_manifest(
            operand['component_name'],
            operand['build_info'],
            image_name_prefix=operand['image_name_prefix'],
            image_remote=operand['image_remote']
        ))
    print(json.dumps(manifest, indent=4))
#    with open('release/image-manifests/1.0.1.json', 'w') as file:
#        file.write(json.dumps(manifest, indent=2))
#        print(manifest)


if __name__ == '__main__':
    main()


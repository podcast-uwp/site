#!/usr/bin/env python3

import argparse
import os
import re
import subprocess
import logging
import sys
import time

from eyed3 import core, id3


def tag_and_upload(args):
    set_mp3_tags(args)
    upload_with_ansible(args)


def upload_with_ansible(args):
    logging.info(f"Running ansible subcommand with args: {args}")
    command = f"ansible-playbook ansible.yml -i {args.host}, -e 'remote_directory={args.dest}' -e " \
              f"'local_file_path={args.file}' -e 'num_days={args.days}'"
    subprocess.call(command, shell=True)


def set_mp3_tags(args):
    """
       Add title, album, artist tags, set album image to podcast episode mp3 file.
       """
    full_path = args.file
    num = get_episode_number(file_path=full_path)

    # remove both ID3 v1.x and v2.x tags.
    remove_version = id3.ID3_ANY_VERSION
    id3.Tag.remove(full_path, remove_version)

    episode_file = core.load(full_path)
    # using ID3v2.3 tags, because using newer ID3v2.4 version leads to problems with Apple Podcasts and Telegram
    # (they will stop showing chapters with long titles at all, see https://github.com/radio-t/radio-t-site/issues/209)
    episode_file.initTag(version=id3.ID3_V2_3)

    tag = episode_file.tag

    try:
        # set mp3 tags
        tag.title = f'{args.title} {num}'
        tag.artist = args.artist
        tag.album = args.album
        tag.track_num = num

        if args.image:
            with open(args.image, "rb") as f:
                image_data = f.read()
                tag.images.set(3, image_data, "image/jpeg")

        tag.save(encoding="utf8")
        print("New mp3 tags are saved.")
    except Exception as exc:
        print("Error:", str(exc), file=sys.stderr)
        sys.exit(1)

    logging.info(f"Setting MP3 tags for file: {args.file}")
    if not os.path.exists(args.file):
        logging.error(f"File does not exist: {args.file}")
        return


def get_episode_number(file_path):
    if not os.path.isfile(file_path):
        raise ValueError(f"File not found: {file_path}")
    match = re.search(r'ump_podcast(\d+)\.mp3', file_path)
    if not match:
        raise ValueError("Invalid file name")
    return int(match.group(1))


def main():
    parser = argparse.ArgumentParser(description='A script with multiple subcommands')
    subparsers = parser.add_subparsers(title='subcommands', dest='subcommand')

    # Common argument group for the subcommands
    common_group = argparse.ArgumentParser(add_help=False)
    common_group.add_argument('--dbg', action='store_true', help='Enable debug logging')
    common_group.add_argument("--file", type=str, required=True, help="the file to operate on")

    # Argument group for run_all subcommand
    run_all_group = argparse.ArgumentParser(add_help=False, parents=[common_group])
    run_all_group.add_argument("--days", type=int, default=120,
                               help="the maximum age of files to keep in the directory")
    run_all_group.add_argument('--host', type=str, default=os.environ.get('HOST', 'localhost'),
                               help='remote server hostname')
    run_all_group.add_argument("--user", type=str, default=os.environ.get('USER', 'umputun'),
                               help="the username to use for the SSH connection")
    run_all_group.add_argument("--dest", type=str, default=os.environ.get('DEST', '/srv/media'),
                               help="the destination directory")
    run_all_group.add_argument("--title", type=str, default=os.environ.get('TITLE', "UWP Выпуск"),
                               help="the title of the MP3 file")
    run_all_group.add_argument("--artist", type=str, default=os.environ.get('ARTIST', 'Umputun'),
                               help="the artist of the MP3 file")
    run_all_group.add_argument("--album", type=str, default=os.environ.get('ALBUM', 'Eженедельный подкаст от Umputun'),
                               help="the album of the MP3 file")
    run_all_group.add_argument("--year", type=str, default=os.environ.get('YEAR', time.strftime("%Y")),
                               help="the year of the MP3 file")
    run_all_group.add_argument("--image", type=str, default=os.environ.get('IMAGE', '/srv/cover.jpg'),
                               help="the path to an image file to use as the album cover")

    # subcommand for uploading files with Ansible
    ansible_parser = subparsers.add_parser("upload", help="upload a file with Ansible", parents=[common_group])
    ansible_parser.set_defaults(func=upload_with_ansible)

    # subcommand for setting MP3 tags
    mp3_parser = subparsers.add_parser("set-mp3-tags", help="set MP3 tags for a file", parents=[common_group])
    mp3_parser.set_defaults(func=set_mp3_tags)

    # subcommand for both tagging and uploading
    run_all_parser = subparsers.add_parser("run-all", help="set MP3 tags and upload a file with Ansible", parents=[run_all_group])
    run_all_parser.set_defaults(func=tag_and_upload)

    args = parser.parse_args()
    if args.dbg:
        logging.basicConfig(level=logging.DEBUG, stream=sys.stdout, format='%(asctime)s.%(msecs)03d %(levelname)s %(message)s',
                            datefmt='%Y-%m-%d %H:%M:%S')
    else:
        logging.basicConfig(level=logging.INFO, stream=sys.stdout, format='%(asctime)s.%(msecs)03d %(levelname)s %(message)s',
                            datefmt='%Y-%m-%d %H:%M:%S')

    if args.subcommand:
        args.func(args)
        return 0

    logging.info(f"Subcommand {args.subcommand} not implemented yet.")
    return 1


if __name__ == '__main__':
    main()

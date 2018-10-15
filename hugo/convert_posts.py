"""
Скрипт для преобразования файлов в формат Hugo
"""

import glob
from collections import defaultdict
from datetime import datetime

SAVE_DIR = './content/posts'
SOURCE_DIR = '../octopress/source/_posts'


def parse_file(name, source, file_counter):
    data = list()
    config_lines = list()
    config_attr = 0

    for line in source:
        if line == '---':
            config_attr += 1
        # эта директива не нужна, так как комментарии закрываются через Disqus
        elif line.startswith('layout') or line.startswith('comments'):
            continue
        elif config_attr == 1:
            config_lines.append(line)
        else:
            data.append(line)

    new_source = ['+++']

    # преобразование настроек
    for config_line in config_lines:
        config_line = config_line.strip()
        if not len(config_line):
            continue

        config = config_line.split(': ', 1)
        if len(config) != 2:
            print('Error (parse config):', config, config_line)
            continue

        key, value = config

        if key == 'date':
            try:
                date = datetime.strptime(value, "%Y-%m-%d %H:%M")
                value = date.strftime('%Y-%m-%dT%H:%M:%S')
            except ValueError:
                print('Error (parse date):', value, name)

        if not value.startswith('"') or not value.endswith('"'):
            if key == 'categories':
                cats = ['"{}"'.format(v) for v in value.split(' ')]
                value = '[{}]'.format(', '.join(cats))
            else:
                value = '"{}"'.format(value.replace('"', '\\"'))

        config_format = '{} = {}'.format(key, value)
        new_source.append(config_format)

    chunks = name.split('-')
    # удаление даты и имени файла
    name = '-'.join(chunks[3:])

    file_counter[name] += 1
    if file_counter[name] != 0:
        new_source.append("aliases = [")
        new_source.append("    \"/p{}/{}/{}/{name}/\",".format(*chunks[:3], name=name))
        new_source.append("    \"/p{}/{}/{}/{name}\"".format(*chunks[:3], name=name))
        new_source.append("]")
        name += str(file_counter[name])

    new_source.append('+++\n')

    # содержание поста
    for line in data:
        new_source.append(line)

    name = SAVE_DIR + '/{}.md'.format(name)
    with open(name, 'w', encoding='utf-8') as h:
        h.write('\n'.join(new_source))

        print(name, 'generated')


def run():
    file_counter = defaultdict(lambda: -1)
    for post_file in glob.glob(SOURCE_DIR + '/*.markdown'):
        name = post_file.replace(SOURCE_DIR, '').replace('.markdown', '').replace('\\', '')
        with open(post_file, encoding='utf-8') as h:
            parse_file(name, h.read().splitlines(), file_counter)


if __name__ == '__main__':
    run()
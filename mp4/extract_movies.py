# -*- coding: utf-8 -*-
import re
import json

# There are 3 components: the title title, the year, and a roman numeral
entry_re = re.compile(r'^(.+)?\s+\(([12?][890?][\d?]{2}(?:/[IVX]+)?)\).*$')


entry_list = []
failed_match_count = 0
entry_count = 0
with open('movies.list', 'rb') as f:
    f.seek(0, 2)
    f_size = f.tell()
    f.seek(0, 0)
    for entry in f:
        entry_count += 1
        entry_match = entry_re.match(entry)
        if entry_match is None:
            print entry
            failed_match_count += 1
            continue
        title, year = entry_match.groups()
        title = unicode(title, errors='ignore')
        entry_list.append((title, year))
        if entry_count % 100000 == 0:
            print '%0.1f%%' % ((float(f.tell()) / f_size) * 100)

print 'Entries:', entry_count
print 'Failed matches:', failed_match_count

entry_index = {}
overused_keywords = set()
for entry in entry_list:
    if entry[0] == '"' and entry[-1] == '"':
        continue
    keyword_title = ''.join([c for c in entry[0].lower() if c.islower() or c == ' ']).strip().split()
    for keyword in keyword_title:
        if len(keyword) < 3 or keyword in overused_keywords:
            continue
        if keyword not in entry_index:
            entry_index[keyword] = []
        entry_index[keyword].append(entry)
        if len(entry_index[keyword]) > 50:
            overused_keywords.add(keyword)
            del entry_index[keyword]

print 'Keywords:', len(entry_index)

with open('entries.json', 'wb') as o:
    json.dump(entry_index, o)

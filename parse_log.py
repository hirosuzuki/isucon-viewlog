#!/usr/bin/python3

import sys
from collections import defaultdict

query_count = defaultdict(int)
query_time = defaultdict(int)
tagquery_count = defaultdict(int)
tagquery_time = defaultdict(int)
tag_count = defaultdict(int)
tag_time = defaultdict(int)

for line in sys.stdin:
    vs = line.strip().split("\t")
    deltatime = int(vs[1])
    if len(vs) == 3:
        tag, query = "", vs[2]
    else:
        tag, query = vs[2], vs[3]
    query_count[query] += 1
    query_time[query] += deltatime
    tagquery_count[(tag, query)] += 1
    tagquery_time[(tag, query)] += deltatime
    tag_count[tag] += 1
    tag_time[tag] += deltatime

def output_table(data, sort_column):
    data.sort(key=lambda x:x[sort_column], reverse=True)
    if isinstance(data[0][3], tuple):
        colnum = 5
        rows = [("total(ms)", "count", "average(ms)", "tag", "content")] + \
            [("%.3f" % (datum[0] / 1000000), str(datum[1]), "%.3f" % (datum[2] / 1000000), datum[3][0], datum[3][1]) for datum in data]
    else:
        colnum = 4
        rows = [("total(ms)", "count", "average(ms)", "content")] + \
            [("%.3f" % (datum[0] / 1000000), str(datum[1]), "%.3f" % (datum[2] / 1000000), datum[3]) for datum in data]

    colws = [max(len(row[i]) for row in rows) for i in range(colnum)]

    formatstr = " | ".join(["%*s"]*3 + ["%-*s"]*(colnum-3))

    def builddata(row):
        data = []
        for i in range(colnum):
            data.append(colws[i])
            data.append(row[i])
        return tuple(data)

    header = formatstr % builddata(rows[0])
    print(header)
    print("-" * len(header))

    for row in rows[1:]:
        print(formatstr % builddata(row))

queries = [(query_time[k], query_count[k], query_time[k] / query_count[k], k) for k in query_count]

print("\n# Top Query (総時間順)\n")
output_table(queries, 0)

print("\n# Top Query (回数順)\n")
output_table(queries, 1)

print("\n# Top Query (平均時間順)\n")
output_table(queries, 2)


tagqueries = [(tagquery_time[k], tagquery_count[k], tagquery_time[k] / tagquery_count[k], k) for k in tagquery_count]

print("\n# Top Tag Query (総時間順)\n")
output_table(tagqueries, 0)

print("\n# Top Tag Query (回数順)\n")
output_table(tagqueries, 1)

print("\n# Top Tag Query (平均時間順)\n")
output_table(tagqueries, 2)

tags = [(tag_time[k], tag_count[k], tag_time[k] / tag_count[k], k) for k in tag_count]

print("\n# Top Tag Query (総時間順)\n")
output_table(tags, 0)

print("\n# Top Tag Query (回数順)\n")
output_table(tags, 1)

print("\n# Top Tag Query (平均時間順)\n")
output_table(tags, 2)


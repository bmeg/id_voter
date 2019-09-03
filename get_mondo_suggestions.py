#!/usr/bin/env python


import sys
import json
import requests


if __name__ == "__main__":
    terms = []
    with open(sys.argv[1]) as handle:
        for line in handle:
            terms.append(line.rstrip())

    for t in terms:
        url = "https://www.ebi.ac.uk/ols/api/search?q=%s&ontology=mondo" % (t)
        h = requests.get(url)
        d = h.json()
        sug = {}
        for r in d['response']['docs']:
            if "obo_id" in r:
                sug[r['obo_id']] = r['label']
        out = {"term" : t, "suggestions" : sug}
        print(json.dumps(out))

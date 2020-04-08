#!/usr/bin/env python
# Copyright (c) 2019 Siemens AG
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
#
# Author(s): Jonas Plum

import json
import os
import sys

import forensicstore
import jinja2
from storeutil import combined_conditions


def transform(store, items, template_name):
    if not items:
        return None
    dir_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), "templates")
    template_loader = jinja2.FileSystemLoader(searchpath=dir_path)
    template_env = jinja2.Environment(loader=template_loader, autoescape=True)

    template = template_env.get_template(template_name)
    output = template.render(data=list(items))

    report_name = template_name.split(".")[0] + ".md"
    with store.store_file("Reports/" + report_name) as (report_path, file_io):
        file_io.write(output.encode('utf-8'))
        return {"type": "report", "report_path": report_path, "format": "markdown"}


def main():
    if len(sys.argv) > 1 and sys.argv[1] == "info":
        print(json.dumps({"Use": "report", "Short": "Generate markdown reports"}))
        sys.exit(0)
    store = forensicstore.connect(".")
    items = list(store.select(sys.argv[1], combined_conditions(None)))
    result = transform(store, items, sys.argv[2])
    if result:
        store.insert(result)
    store.close()


if __name__ == '__main__':
    main()

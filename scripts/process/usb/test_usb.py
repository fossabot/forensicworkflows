import os
import shutil
import tempfile

import forensicstore
import pytest
from .usb import main


@pytest.fixture
def data():
    tmpdir = tempfile.mkdtemp()
    shutil.copytree("test", os.path.join(tmpdir, "data"))
    return os.path.join(tmpdir, "data")


def test_usb(data):
    cwd = os.getcwd()
    os.chdir(os.path.join(data, "data", "usb.forensicstore"))

    main()

    store = forensicstore.connect(os.path.join(data, "data", "usb.forensicstore"))
    items = list(store.select("usb-device"))
    store.close()
    assert len(items) == 1

    os.chdir(cwd)
    shutil.rmtree(data)

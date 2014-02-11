__author__ = 'imranp'

import sys
import multiprocessing
from multiprocessing import Value
import uuid
from hashlib import sha1

def githash(data):
    s = sha1()
    s.update("commit %u\0%s" % (len(data), data))
    return s.hexdigest()

def worker(commit_prefix, difficulty, id, done):
    i = 0
    while done.value == 0:
        nonce = str(id) + "-" + str(i)
        commit = commit_prefix + nonce + "\n"
        i += 1
        if i % 100000 == 0:
            sys.stderr.write('.')
            sys.stderr.flush()
        hash = githash(commit)
        if hash < difficulty and done.value == 0:
            done.value = 1
            sys.stdout.write(nonce)
            sys.stderr.write(hash + " " + str(len(commit)) + "\n")
            break

if __name__ == '__main__':
    commit = sys.argv[1]
    difficulty = sys.argv[2]
    n = int(sys.argv[3])
    done = Value('i', 0)
    for i in range(n):
        id = uuid.uuid4()
        p = multiprocessing.Process(target=worker, args=(commit,difficulty,id,done))
        p.start()

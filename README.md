## Stripe CTF3 Solutions

My solutions for [Stripe's Capture the Flag](https://stripe-ctf.com/). The resulting scores are [here](https://stripe-ctf.com/achievements/isb). Here is a brief description of how each solution works:

### level 0
Simply using a Ruby set to store the dictionary is enough to clear this level. Lots of optimizations are possible - for e.g. loading an offline-generated dictionary index.

### level 1
Modified the supplied miner script to fork a python script to do the hashing. The Python script (gitcoin-miner.py) uses the multiprocessing module to parallelize hashing.

### level 2
Simply allowing one concurrent connection at a time from each unique IP address was good enough to block most elephants. This is an interesting problem but sadly I didn't have enough time to work on it!

### level 3
I spent a fair amount of time trying different indexes to improve my score. However, the first two simple changes I made gave me the best score:

1.  Shard the index across the servers based on document id. 
2.  Instead of reading files from disk at query time, build a line-based in memory index.

### level 4
This was the most painful level to get working and there is a fair bunch of copy-paste code. These were the major changes:

1.  Integrate the [goraft](https://github.com/goraft/) library by following the raftd reference implementation.
2.  Change the provide solution to use unix transport instead of http.
3.  Implement leader forwarding to route queries to the current leader.
4.  Use an in-memory sqlite for the database to speed up queries/updates.
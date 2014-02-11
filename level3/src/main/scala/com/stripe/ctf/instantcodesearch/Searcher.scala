package com.stripe.ctf.instantcodesearch

import java.io._
import java.nio.file._

import com.twitter.concurrent.Broker

abstract class SearchResult
case class Match(path: String, line: Int) extends SearchResult
case class Done() extends SearchResult

class Searcher(index : Index)  {

  def search(needle : String, b : Broker[SearchResult]) = {
    val startTime = System.currentTimeMillis();

    for ((word, pos) <- index.words) {
      if (word.contains(needle)) {
        pos.foreach{ case (docId, line) => b !! new Match(index.docs(docId), line) }
      }
    }
    System.err.println(
      needle + " took " +
      (System.currentTimeMillis() - startTime)  + " ms"
    )

    b !! new Done()
  }

}

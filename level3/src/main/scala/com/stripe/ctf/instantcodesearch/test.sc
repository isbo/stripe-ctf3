package com.stripe.ctf.instantcodesearch
import scala.collection.mutable._

object test {
  println("Welcome to the Scala worksheet")       //> Welcome to the Scala worksheet
  
    def tokenize(s: String): MutableList[String] = {
        var words = MutableList[String]()
        var start = 0
    	for((c,i) <- s.view.zipWithIndex) {
    	  if (c == ' ') {
    		  if (i > start) words += s.slice(start, i)
    		  start = i + 1
    	  }
    	}
      if (s.length > start) words += s.slice(start, s.length)
        words
    }                                             //> tokenize: (s: String)scala.collection.mutable.MutableList[String]

	 tokenize("a b cd e")                     //> res0: scala.collection.mutable.MutableList[String] = MutableList(a, b, cd, e
                                                  //| )

}
package com.stripe.ctf.instantcodesearch

import java.io._
import scala.collection.mutable._
import java.nio.file.Path
import scala.io.Source
import java.nio.charset.Charset
import java.nio.file.Files
import java.nio.charset.CodingErrorAction

class Index(root: Path) extends Serializable {
  var words = HashMap[String, MutableList[(Int, Int)]]()
  var docs = MutableList[String]()
  var docId = -1
  val decoder = Charset.forName("UTF-8").newDecoder()
  decoder onMalformedInput CodingErrorAction.REPORT
  decoder onUnmappableCharacter CodingErrorAction.REPORT
  
  def addFile(filePath: Path) {    
    docs += root.relativize(filePath).toString
    docId += 1

    /*
    val bytes = Files.readAllBytes(filePath)
    val r = new InputStreamReader(new ByteArrayInputStream(bytes), decoder)
    val strContents = slurp(r)

    strContents.split("\n").zipWithIndex.foreach {
      case (l,n) => if (!words.contains(l)) words(l) = MutableList[(Int, Int)]()
        words(l) += ((docId, n+1))
    }

    */
        
    Source.fromFile(filePath.toString()).getLines().zipWithIndex.foreach {
      case (l,n) => if (!words.contains(l)) words(l) = MutableList[(Int, Int)]()
        words(l) += ((docId, n+1))
    }
    
  }
}


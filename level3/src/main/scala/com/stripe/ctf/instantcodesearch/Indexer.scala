package com.stripe.ctf.instantcodesearch

import java.io._
import java.util.Arrays
import java.nio.file._
import java.nio.charset._
import java.nio.file.attribute.BasicFileAttributes

class Indexer(indexPath: String, id: Int) {
  val root = FileSystems.getDefault().getPath(indexPath)
  
  def index() : Index = {
	val idx = new Index(root)
    var total = 0
    var indexed = 0
    val startTime = System.currentTimeMillis();
    Files.walkFileTree(root, new SimpleFileVisitor[Path] {
      override def preVisitDirectory(dir : Path, attrs : BasicFileAttributes) : FileVisitResult = {
        if (Files.isHidden(dir) && dir.toString != ".")
          return FileVisitResult.SKIP_SUBTREE
        return FileVisitResult.CONTINUE
      }
      override def visitFile(file : Path, attrs : BasicFileAttributes) : FileVisitResult = {
        if (Files.isHidden(file))
          return FileVisitResult.CONTINUE
        if (!Files.isRegularFile(file, LinkOption.NOFOLLOW_LINKS))
          return FileVisitResult.CONTINUE
        if (Files.size(file) > (1 << 20))
          return FileVisitResult.CONTINUE
        /*
        if (Arrays.asList(bytes).indexOf(0) > 0)
          return FileVisitResult.CONTINUE
        */
        try {
          val node_id = (math.abs(file.hashCode()) % 3) + 1 // yuck
          if (node_id == id) {
            idx.addFile(file)
            indexed += 1
          }
          total += 1
        } catch {
          case e: IOException => {
            return FileVisitResult.CONTINUE
          }
        }

        return FileVisitResult.CONTINUE
      }
    })
    System.err.println(
      "node " + id + " indexed " + indexed + " documents out of " + total + " in " +
      (System.currentTimeMillis() - startTime) + " ms"
    )

    return idx
  }

}

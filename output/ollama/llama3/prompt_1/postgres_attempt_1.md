As we continue our blog series comparing Git and Dolt, this week we're diving into three fundamental commands that every developer knows: `clone`, `pull`, and `push`. These commands are essential for managing code repositories in Git, but what about Dolt? Let's explore how these commands work in both tools and highlight the key differences.

### The Basics

Before we dive into the specifics, let's cover the basic purposes of each command:

* `clone`: Creates a local copy of a repository from a remote server.
* `pull`: Fetches changes from a remote repository and merges them with your local version.
* `push`: Updates a remote repository with your local changes.

### Git: The Original

In Git, these commands are used to manage code repositories. Let's take a look at some terminal-based examples:

```
$ git clone https://github.com/dolthub/dolt.git
Cloning into 'dolt'...
remote: Counting objects: 100% (3/3), done.
remote: Compressing objects: 100% (2/2), done.
Receiving object: 100% (3/3), 243.00 KiB | 1.42 MiB/s, done.
$ git pull origin master
remote: Counting objects: 100% (5/5), done.
remote: Compressing objects: 100% (4/4), done.
Receiving object: 100% (5/5), 12.44 KiB | 1.41 MiB/s, done.
Merge made by the book.
$ git push origin master
Counting objects: 100% (3/3), done.
Writing objects: 100% (3/3), 243.00 KiB | 1.42 MiB/s, done.
```

### Dolt: The SQL Database

In Dolt, these commands are used to manage database versions. Here's how they work:

```
$ dolt clone https://github.com/dolthub/dolt.git
Cloning into 'dolt'...
remote: Counting objects: 100% (3/3), done.
remote: Compressing objects: 100% (2/2), done.
Receiving object: 100% (3/3), 243.00 KiB | 1.42 MiB/s, done.
$ dolt pull origin master
remote: Counting objects: 100% (5/5), done.
remote: Compressing objects: 100% (4/4), done.
Receiving object: 100% (5/5), 12.44 KiB | 1.41 MiB/s, done.
Merge made by the book.
$ dolt push origin master
Counting objects: 100% (3/3), done.
Writing objects: 100% (3/3), 243.00 KiB | 1.42 MiB/s, done.
```

### Key Differences

The most significant difference between Git and Dolt is the type of data being managed. Git is designed for versioning code, while Dolt is designed for managing database versions. This fundamental difference affects how each command works.

For example, when you use `git clone`, you're creating a local copy of a repository that contains code files. In Dolt, `clone` creates a local copy of a database that contains schema and data. Similarly, when you use `git pull`, you're fetching changes from a remote repository and merging them with your local version. In Dolt, `pull` fetches changes from a remote repository and merges them with your local database.

### Why Dolt Matters

Dolt's first-class versioning features, built on top of Git-like semantics, make it an exciting technology for data versioning. By understanding how these commands work in both tools, you can better appreciate the transformative power of Dolt.

Next week, we'll be exploring another fundamental command: `commit`. Stay tuned!

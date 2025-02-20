**Comparing `clone`, `pull`, and `push` in Git vs. Dolt**

Dolt is an OLTP SQL database that was built from the ground up with versioning, branching, and merging. These first-class versioning features use Git-like semantics, making Dolt immediately familiar to Git users. In this post, we'll compare and contrast how `clone`, `pull`, and `push` work in Git vs. Dolt.

### How it works

In Git, `clone` is used to create a local copy of a remote repository. This command is useful when you want to start working on a project from scratch or when you want to duplicate an existing repository for testing or backup purposes.

In Dolt, `clone` is similar to its Git counterpart. When you run `dolt clone <repo>`, Dolt creates a local copy of the specified repository. This command is useful when you want to start working on a project from scratch or when you want to duplicate an existing repository for testing or backup purposes.

In Git, `pull` is used to fetch and merge changes from a remote repository into your local repository. This command is useful when you want to update your local repository with the latest changes from the remote repository.

In Dolt, `pull` is similar to its Git counterpart. When you run `dolt pull <repo>`, Dolt fetches and merges changes from the specified repository into your local repository. This command is useful when you want to update your local repository with the latest changes from the remote repository.

In Git, `push` is used to upload changes from your local repository to a remote repository. This command is useful when you want to share your changes with others or when you want to backup your work.

In Dolt, `push` is similar to its Git counterpart. When you run `dolt push <repo>`, Dolt uploads changes from your local repository to the specified repository. This command is useful when you want to share your changes with others or when you want to backup your work.

### Key differences

One key difference between Git and Dolt is that Dolt uses a different underlying database technology than Git. While Git stores its data in a flat file system, Dolt stores its data in a relational database. This means that Dolt can take advantage of the ACID properties of a relational database, which provides better support for concurrent transactions.

Another key difference between Git and Dolt is that Dolt has first-class support for versioning, branching, and merging. While Git also supports these features, they are not as tightly integrated into the core product as they are in Dolt.

### Why Dolt is a transformative technology

Dolt's ability to integrate versioning, branching, and merging with a relational database makes it a transformative technology for data versioning. By providing a single platform for both data storage and versioning, Dolt simplifies the process of managing complex data sets and reduces the risk of errors.

In conclusion, while Git is a powerful tool for software versioning, Dolt provides a more comprehensive solution for data versioning. By integrating versioning, branching, and merging with a relational database, Dolt provides a single platform for both data storage and versioning, making it easier to manage complex data sets and reduce the risk of errors.

**Teaser for next post**

In our next blog post, we'll be exploring how Dolt's support for concurrent transactions can help you build more scalable and reliable applications. We'll also be introducing some new features that make it even easier to work with large datasets in Dolt. Stay tuned!
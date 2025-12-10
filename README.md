## blog

build a very basic blog in Java using Spring Boot, Spring Web, Spring Data JPA, and an in-memory H2 database.

## Directory Structure

```
src/
 └── main/
      ├── java/com/example/blog/
      │       ├── BlogApplication.java
      │       ├── controller/
      │       │       └── PostController.java
      │       ├── model/
      │       │       └── Post.java
      │       └── repository/
      │               └── PostRepository.java
      └── resources/
              ├── templates/
              │       ├── index.html
              │       └── post.html
              └── application.properties
```

## Run the Application

If using Maven:
```
mvn spring-boot:run
```
Then open: http://localhost:8080

## Done!

You now have a fully working minimal blog built in Java with Spring Boot.

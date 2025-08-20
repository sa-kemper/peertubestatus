# golangGetTextTest
This is a simple golang webserver that localizes it's pages to the requested language (if available) in a concurrently safe manner, this project is used mainly for testing purposes but may be used as a starting point of a real project.

By default, the Application embeds every asset. you can override each deployed asset at runtime.

---
### Override content.
| Override file type      | override location                                       |
|-------------------------|---------------------------------------------------------|
| Override html templates | .gohtml files into the work directory/TemplateOverride/ |

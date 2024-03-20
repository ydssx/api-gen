## API-GEN

API-GEN is a tool that generates APIs and handlers based on a configuration file and predefined type structures. It simplifies the process of creating APIs by automatically generating the necessary code.

### Getting Started

1. To install API-Gen, use the following command::

```
go install github.com/ydssx/api-gen@latest
```

2. Define the type structures in the `typeFile`:

In the `typeFile`, you need to define the type structures for your APIs. Each structure should be annotated with metadata comments that specify the group, authentication requirement, handler, and router details.

Example:

```go
// @group apiv22
// @auth false
// @handler register
// @router /register [get]
type (
	RegisterReq struct {
		Name     string `form:"name" binding:"required"` //用户名
		Password string `form:"password"`
	}

	RegisterResp struct {
		User string `json:"user"`
	}
)
```

3. Configure the `config.yaml` file:

The `config.yaml` file contains the configuration settings for API-GEN. You can specify the API paths, type file path, logic file, handler file, and router file.

Example:

```yaml
apiPath:
  - /register
  - /login

typeFile: example/types/example.go

logic:
  file: example/logic/logic.go
  receiver: "*UserLogic"

handler:
  file: example/handler/handler.go

router:
  file: example/router/router.go
  groupFunc: UserRouter

```

4. Run the API-GEN tool:

To generate the APIs and handlers based on the configuration and type structures, run the following command:

```
api-gen -c config.yaml
```

This will read the `config.yaml` file, parse the type structures from the `typeFile`, generate logic functions, handler functions, and add routers accordingly.

### Configuration Options

- `apiPath`: A list of API paths where the generated APIs will be registered.
- `typeFile`: The path to the file that contains the type structures for the APIs.
- `logic.file`: The file where the logic functions will be generated.
- `logic.receiver`: The receiver name for the logic functions.
- `handler.file`: The file where the handler functions will be generated.
- `router.file`: The file where the router functions will be generated.
- `router.groupFunc`: The name of the group function in the router file.

### Generated Files

API-GEN generates the following files based on the configuration:

- Logic file: Contains the logic functions for the APIs.
- Handler file: Contains the handler functions for the APIs.
- Router file: Contains the router functions that register the APIs.

### Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the API-GEN repository.

### License

This project is licensed under the [MIT License](LICENSE).

## Changelog

### v0.1.0

- Initial release.

### v0.2.0

- Add support for multiple API paths.
- Add support for multiple type files.
- Add support for multiple logic files.
- Add support for multiple handler files.
- Add support for multiple router files.
- Add support for multiple group functions in the router file.
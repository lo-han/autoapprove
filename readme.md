How to use:

- Create a .env file with your [Github Personal Access Token](https://github.com/settings/tokens)

```
TOKEN=<your_token>
```

- Run the app (watch mode):

```
export $(cat .env | xargs) && go run . watch
```

- Run the app (cli mode):

```
export $(cat .env | xargs) && go run . cli <pull_request_1> <pull_request_2> ...
```


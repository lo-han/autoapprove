How to use:

- Create a .env file with your [Github Personal Access Token](https://github.com/settings/tokens)

```
TOKEN=<your_token>
```

- Run the app:

```
export $(cat .env | xargs) && go run . <pull_request_1> <pull_request_2> ...
```

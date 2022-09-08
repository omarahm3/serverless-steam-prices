# Steam games prices

![image](https://user-images.githubusercontent.com/8606113/189094074-3221c2d4-10f4-4b24-a116-2eef4ee3d455.png)

Serverless playground project, the main goal is to look for games on steam via steam API and show some details off of it for example: price, image.

## Run

```bash
make build
# Using netlify CLI to run the function
ntl functions:serve
# Install frontend packages then run the dev server
yarn --cwd frontend/
yarn --cwd frontend/ dev
```

## TODO

- [ ] Deploy function to lambda and S3 using serverless framework
- [ ] Cache steam requests especially all apps request (Redis, or mongo maybe?)
- [ ] Show proton Linux support for games

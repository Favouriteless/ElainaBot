docker build . -t elaina:latest

docker image save elaina:latest -o build/image
scp .\build\image favouriteless@maven.favouriteless.net:/srv/favouriteless-maven/image
DEL /F /Q build
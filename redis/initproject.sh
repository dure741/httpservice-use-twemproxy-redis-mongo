NOWPP=$(pwd)

## set GOPATH
if [ "$GOPATH" = "" ]
then
    echo please set GOPATH
    exit 127
fi

#install package golang.org/x/net
cd dep_pack
tar zxvf golang.org.x.net.tar.gz
mv golang.org $GOPATH/src
cd ..

# install dep
go get -v -u github.com/golang/dep/cmd/dep


## get iris
#go get github.com/kataras/iris@v11.1.1

mkdir -p $GOPATH/src/github.com/kataras

wget https://github.com/kataras/iris/archive/v11.1.1.tar.gz
tar zxvf v11.1.1.tar.gz
mv iris-11.1.1 iris
mv iris $GOPATH/src/github.com/kataras

project=$1
if [ "$project" = "" ]
then
    echo project name is empty, usage shell project_name
    exit 127
fi

myos=$(uname)
case "$myos" in
    "Darwin")
        #sed -i "" 's#redis#'$project'#g' `grep redis * -rl`
        find . -type f -name '*.go' -exec sed -i '' s/redis/$project/ {} +
        sed -i "" 's#hello#'$project'#g' cmd/redis/main.go
        ;;
    "Linux")
        sed -i 's#redis#'$project'#g' `grep redis * -rl`
        sed -i 's#hello#'$project'#g' cmd/redis/main.go
        ;;
    *)
        echo not support
        exit
        ;;
esac

mv cmd/redis cmd/$project
cd ..
[ -d "redis" ] && mv redis $project
[ -d "HeraGo" ] && mv HeraGo $project

cd $project
rm -rf .git
echo $project is ok


#######
#
#  1. if your OS is windows, Please manually replace 'redis' with your project name , Including in files and paths
#

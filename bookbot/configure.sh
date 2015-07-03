
upst() {
    sed -i "s/yourusername/$(whoami)/g" etc/bookbot.conf
    sed -i "s:bookbotpath:$(pwd):g" etc/bookbot.conf
}


directory() {
    echo "Type the base directory"
    read bookspath
    if [ ! -d "`eval echo ${bookspath//>}`" ];then
        echo "Base directory $bookspath not valid"
        exit 1
    fi
}

database() {
    echo "Type the data base directory path"
    read dbpath
    if [ -f "`eval echo ${dbpath//>}`" ];then
        echo "Base directory $dbpath not valid"
        exit 1
    fi
}

environ() {
    echo "Type the secrets file path"
    read envpath
    if [ ! -f "`eval echo ${envpath//>}`" ];then
        echo "Secrets file path $envpath not valid"
        exit 1
    fi

}

pass() {
    echo "Type the password you want (optional)"
    echo -n "Password: "
    read -s password
    echo
    echo -n "Password(again): "
    read -s password2
    echo

    if [ ! $password == $password2 ];then
        echo "Your passwords don't math"
        exit 1
    fi
}

all() {
    upst
    directory
    database
    environ
    pass
    echo "I'm gonna configure with the following values:"
    echo "--------"
    echo "Base path: $bookspath"
    echo "DB path: $dbpath"
    echo "Secrets path: $envpath"
    echo "The password you writed (if some)"
    echo "--------"
    echo -n "Do you want to continue? (y/N): "
    read ok
    if [ ! $ok == "y" ];then
        echo "Not writing the file"
        exit 1
    fi

    sed -i "s:BOOKSPATH=.*:BOOKSPATH=$bookspath:g" run.sh
    sed -i "s:DBPATH=.*:DBPATH=$dbpath:g" run.sh
    sed -i "s:ENVDIR=.*:ENVDIR=$envpath:g" run.sh
    sed -i "s:PWD=.*:PWD=$password:g" run.sh
}

if [ "$1" = "all" ];then
    all
fi

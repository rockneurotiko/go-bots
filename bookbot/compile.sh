cd bin
gox ..
cp ../secrets.env .
sed -i "s/TELEGRAM_KEY=.*/TELEGRAM_KEY=YOURKEY/" secrets.env

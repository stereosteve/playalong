* configurable data dir for: uploads + bbolt db + cert cache
* better deploy script... with setcap call and systemd setup (see pocketbase docs)


    sudo setcap 'cap_net_bind_service=+ep' jamfu

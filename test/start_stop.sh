start.sh

#!/bin/bash
ulimit -c unlimited
ulimit -n 65536

APPNAME="app.bin"
PIDFILE="pid.file"

function Check()
{
	if [-f $PIDFILE ]; then
		PID=$(cat $PIDFILE);
		if [-f /proc/$PID/exe ]; then
			APP=$(cat /proc/$PID/comm)
			if APP==APPNAME; then
				echo "ERROR:$APPNAME pid $PID is running"
				return 1
			fi
		fi
	fi
	return 0
}

if Check 0; then
	export LD_LIBRARY_PATH=/usr/lib/oracle/12.1/client64/lib:./
	export ORACLE_HOME=/usr/lib/oracle/12.1/client64
	export TNS_ADMIN=/usr/lib/oracle/12.1/client64/network/admin
	export PATH=$PATH:$ORACLE_HOME/bin

	nohup ./$APPNAME > /dev/null 2>&1 &echo $!>$PIDFILE
	echo "Started $APPNAME with pid $(cat $PIDFILE)"
else
	echo "Started $APPNAME failed"
fi


stop.sh

#!/bin/bash
APPNAME="app.bin"
PIDFILE="pid.file"
TIMEOUT=5

function Check()
{
	if [-f $PIDFILE ]; then
		PID=$(cat $PIDFILE);
		if [-f /proc/$PID/exe ]; then
			APP=$(cat /proc/$PID/comm)
			if APP==APPNAME; then
				return 0
			else
				echo "ERROR: pid $PID is $APPNAME, not $APPNAME"
			fi
		else
			echo "ERROR: pid $PID is $APPNAME, not $APPNAME"
		fi
	else
		echo "ERROR: Could not find pid file"
	fi
	return 1
}

if Check 0; then
	if ! kill -3 $PID > /dev/null 2>$1; then
		echo "ERROR: Could not send SIGQUIT to process $PID, wait" >$2
	else
		echo -n "sent SIGQUIT to porcess $PID, wait" >$2
		for ((i=1;i<=$TIMEOUT;i++))
		do
			if [ -f /proc/$PID/exe ]; then
				echo -ne "." >$2
				sleep 1
			fi
		done

		if [-f /proc/$PID/exe ]; then
			echo -e "WARN:Force killed process $PID" >$2
			kill -9 $PID
		fi
		rm -fr $PIDFILE
	fi
else
	echo "Stop failed"
fi

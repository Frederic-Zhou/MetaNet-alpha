import sys
import os


for line in sys.stdin:
    with open("./events.log", 'a') as f:

        event_type = os.environ.get('SERF_USER_EVENT') if os.environ.get(
            'SERF_USER_EVENT') else os.environ.get('SERF_QUERY_NAME')

        if event_type is None:
            event_type = ""

        f.write("%s %s %s:%s" %
                (os.environ.get('SERF_EVENT'), os.environ.get('SERF_SELF_NAME'), event_type, line))

print('copied that!')


# echo from $SERF_SELF_NAME $SERF_EVENT $SERF_USER_EVENT$SERF_QUERY_NAME:$(</dev/stdin) >>events.log ;echo  COPIED THAT!

import thread,time,Queue


i = 0
mutChan = Queue(1)

def main():

	
	global i
	mutChan.put(True)
	thread.start_new_thread(thread1,())
	thread.start_new_thread(thread2,())
	time.sleep(10)
	print "nubmer", i


def thread1():
	global i
	for j in range(1,1000000):
		i+=1

def thread2():
	global i
	for j in range (1,1000000):
		i-=1

main()
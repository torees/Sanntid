import thread,time,Queue,threading
#hallo

i = 0
#lock = threading.Lock()
queLock = Queue.Queue(1)

def main():


	global i
	thread.start_new_thread(thread1,())
	thread.start_new_thread(thread2,())
	time.sleep(10)
	print "nubmer", i


def thread1():
	#lock.acquire()
	queLock.put(1)
	global i
	for j in range(1,1000000):
		i+=1
	#lock.release()
	queLock.get()

def thread2():
	queLock.put(1)
	#lock.acquire()
	global i
	for j in range (1,1000000):
		i-=1
	#lock.release()
	queLock.get()
main()

import thread,time


i = 0

def main():

	
	global i
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
#include <pthread.h>
#include <stdio.h>


int j = 0;

void *thread1(){
	for (int i = 0; i < 1000000; i++){
		j++;
	}
	return 0;
}

void *thread2(){
	for (int i = 0; i < 1000000; i++){
		j--;
	}
	return 0;
}

int main(){

	pthread_t someThread1;
	pthread_t someThread2;

	pthread_create(&someThread1, NULL, thread1,NULL);
	pthread_join(someThread1, NULL);
	
	pthread_create(&someThread2, NULL, thread2,NULL);
	pthread_join(someThread2, NULL);



	printf("%d \n", j);

	return 0;
}

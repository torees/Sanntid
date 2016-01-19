#include <pthread.h>
#include <stdio.h>
#include <time.h>


int j,i = 0;
pthread_mutex_t mutex_j = PTHREAD_MUTEX_INITIALIZER;

void *thread1(){
	pthread_mutex_lock(&mutex_j);
	for (i = 0; i < 1000000; i++){
		j++;
	}
	pthread_mutex_unlock(&mutex_j);
	return 0;
}

void *thread2(){
	pthread_mutex_lock(&mutex_j);
	for (i = 0; i < 100000; i++){
		j--;
	}
	pthread_mutex_unlock(&mutex_j);
	return 0;
}

int main(){

	pthread_t someThread1;
	pthread_t someThread2;

	pthread_create(&someThread1, NULL, thread1,NULL);
	pthread_join(someThread1, NULL);
	
	pthread_create(&someThread2, NULL, thread2,NULL);
	pthread_join(someThread2, NULL);


	pthread_mutex_destroy(&mutex_j);
	printf("%d \n", j);

	return 0;
}

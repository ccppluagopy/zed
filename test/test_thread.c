#include <stdio.h>
#include <sys/syscall.h>
#include <pthread.h>
#include <time.h>

void func(void *arg){
  printf("pthread_self(): %d\n", pthread_self());
  printf("syscall(SYS_gettid): %d\n", syscall(SYS_gettid));
}

int main(){
  pthread_t tt1 = pthread_self();
  int tt2 = syscall(SYS_gettid);
  int t1 = time(0);
  long long i = 0;
  long long loop = 20000000;
  int tn1 = 0;
  int tn2 = 0;
  for(i=0; i<loop; i++){
    if(pthread_equal(tt1, pthread_self())){
       tn1++;
    }
  }
  int t2 = 0;
  for(i=0; i<loop; i++){
    if(tt2 == syscall(SYS_gettid)){
       tn2++;
    }
  }
  int t3 = time(0);
  printf("tn1: %d, tn2: %d\n", tn1, tn2);
  printf("t2-t1: %d t3-t2: %d\n", t2-t1, t3-t2);
  pthread_create(&tid, 0, fun, 0);
  pthread_t tid;
  pthread_join(tid, 0);
  getchar();
  return 0;
}

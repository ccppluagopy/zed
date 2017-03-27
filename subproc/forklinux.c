#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/wait.h>

int main(int argc, char **argv)
{
  if(argc == 0)
  {
    printf("forklinux error: 0 args, need at least 1 args.\n";
    return 0;
  }
  
  int pid;
  if(pid = fork() > 0)
  {
    waitpid(pid, 0, 0);
  }else{
    int ret = execv(argv[0], argv+1);
    printf("exec error, ret: %d, %m\n", ret);
  }
  
  return 0;
}

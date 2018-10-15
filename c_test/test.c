#include <stdio.h>
#include <stdlib.h>

int main() {
  char *p = "32+4";
  printf("%ld\n", strtol(p, &p, 10));
  return 0;
}

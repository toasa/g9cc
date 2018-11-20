#include <stdio.h>

void printBoard(int board[100]) {
  for (int i = 0; i < 9; i=i+1) {
	for (int j = 0; j < 9; j=j+1) {
	  if (board[i * 9 + j] == 0) {
		printf(". ");
	  } else {
		printf("Q ");
	  }
	}
	printf("\n");
  }
  printf("\n");
}

int isQ(int board[100], int row, int col) {
  if (board[9*row + col] == 1) {
	return 1;
  } else {
	return 0;
  }
}

int conflict(int board[100], int row, int col) {
  for (int i = 1; i < 9; i++) {
	// check diagonal
	if ((row-i) >= 0 && (col-i) >= 0) {
	  if (isQ(board, row-i, col-i)) {
		goto conf;
	  }
	}
	if ((row-i) >= 0 && (col+i) < 9) {
	  if (isQ(board, row-i, col+i)) {
		goto conf;
	  }
	}
	if ((row+i) < 9 && (col-i) >= 0) {
	  if (isQ(board, row+i, col-i)) {
		goto conf;
	  }
	}
	if ((row+i) < 9 && (col+i) < 9) {
	  if (isQ(board, row+i, col+i)) {
		goto conf;
	  }
	}

	for (int j = 1; j < 9; j++) {
	  // check vertical
	  if (row - i >= 0) {
		if (isQ(board, row-i, col)) {
		  goto conf;
		}
	  }
	  if (row + i < 9) {
		if (isQ(board, row+i, col)) {
		  goto conf;
		}
	  }

	  // check horizontal
	  if ((col - j) >= 0) {
		if (isQ(board, row, col-j)) {
		  goto conf;
		}
	  }
	  if ((col + j) < 9) {
		if (isQ(board, row, col+j)) {
		  goto conf;
		}
	  }
	}
  }
  return 0;
  
 conf:
  return 1;
}


void solve(int board[100], int row) {
  if (row > 9) {
	printBoard(board);
	return;
  }
  for (int col = 0; col < 9; col++) {
	if (conflict(board, row, col)) {
	  continue;
	} else {
	  board[9*row + col] = 1;
	  solve(board, row+1);
	  board[9*row + col] = 0; // I forget this stmt ><
	}
  }
}

int main(){
  int board[100];
  for (int i = 0; i < 100; i=i+1) {
	board[i] = 0;
  }
  solve(board, 0);
  return 0;
}

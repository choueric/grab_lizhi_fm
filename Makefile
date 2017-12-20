EXEC = $(shell basename ${PWD})

$(EXEC): *.go
	@go build -o $(EXEC) -v

clean:
	@rm -rfv $(EXEC)

run: $(EXEC)
	./$(EXEC)

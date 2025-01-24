metricstest -test.v -test.run="^TestIteration14[AB]*$" ^
-binary-path=../server/server.exe -source-path=../server/ ^
-agent-binary-path=../agent/agent.exe ^
-server-port=8080 -file-storage-path=goshran.txt ^
-database-dsn=postgres://postgres:passwordas@localhost:5432/forgo ^
-key=pass
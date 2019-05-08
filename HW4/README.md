# Machine Problem 4 - Hole Punch

This project is an Hole Punch service implemented in Go, whose lightweight goroutine feature is good fit for the multi-threading requirement of this work. The Hole Punch aims to help unaccessible local services communicate with external clients.

The Hole Punch includes three parts: server, client, and config. The public accessible punch server receives external connections for the actual services and match these connections with the received punch client connections initiated by the punch client running in the local as well as the actual service. The punch client initiates a pair of connections to the punch server and the actual service respectively for every incoming connection to the punch server and exchange data between the pair. The config is a command-line interface to register punch client user accounts.

## Usage

The punch server should run in a server with available public IP and given the control port to listen for punch clients.

```
./punchserver [control port]
```

The punch client should be started within the same machine of the local service. Its mandatory arguments include the port of the local service, the punchserver address, client username, password, and the remote port for external access.

```
./punchclient [local port] [punch server addr] [username] [password] [remote port]
```

The punch client can also be used to list the existing transactions of a given punch server.

```
./punchclient LIST [punch server addr]
```

The punch config needs to run in the same machine with the punch server. It writes new client credentials to a hardcoded path at `HOME/.hole_punch/users`.

```
./punchconfig
```


## Implementation

### Pipe Connections

Exchanging the data stream of two connections is an essential utility required by both the punch server and the punch client. This util function receives two connections and optionally two corresponding message channels. For each connection, it starts a new thread to read from the connection and write the contents to the other one. Overall, it pipes any data between two connections in both directions. If the optional message channel is given, it also sends the number of bytes read from the connection to the channel right after each write, so that other threads can listen at the other end of the channel to get the realtime updates of the amount of data streamed, which supports the "LIST" command in the punch server. The function itself is blocking and it waits until either of the threads ends for exceptions like connection close.

### Punch Config

The punch config creates the file at local path `HOME/.hole_punch/users` if it does not exist and reads all the contents from it. It parses the contents to a list of user object structure which consists of the username, password, and port. Then the process scans from stdin for the three fields in sequence to create a new user item. It verifies every input and asks for re-input if the input is invalid. Username cannot be empty, contain any spaces or duplicate with any existing ones in the user list. Password cannot empty or contain any spaces. Port must be able to be parsed to a number and cannot duplicate with the ports of the existing list. The password is hashed with a random salt. The new user structure is appened to the file.


### Punch Server

The main process of the punch server listens to the port specified in the arguements. For each incoming connection, the process starts a new thread to read it. The read data is converted into a string and treated as the command. The supported commands include "OPEN" for opening a port for the punch client and "LIST" for replying the status of all the current open connections. The punch server only reads from this connection for command once and ignores any incoming data.

For "OPEN" command, the punch server verifies the provided username, password, and port with the configuration file. For an invalid credential, the server writes "FAIL" back to the client and closes the connection. For a valid credential, the server keeps this connection as the control connection for this client session. Then the server listens to the requested port while it randomly opens another port and saves it as the cliennt facing port for this seesion. The server saves a record of this client session which is used in "LIST" command to list the open connections. For every incoming external connection, the server sennds a generated nounce with the client facing port through the control connection to invite the punch client initiate another connection. Meanwhile, it listens to the client facing port in another thread which returns the received client connection with matching nounce or an error otherwise. The server waits for the client connection thread to join with a ten-second timeout. If it times out or returns error, the server closes the client session. If it returns a client connection, the server creates another pipe thread to handle the pair of external and client connections and go back to wait for the next external connection. The pipe thread then takes use of the pipe connections utility to exchange data between the external and client connections with two corresponding message channels. The pipe thread listens to the channels to accumulate the bytes tranferred into the session record in real time. The pipe thread terminates when the either of the connection closes. While the server closes the client session, it closes the outward facing port and client facing port, removes the open connection record, and terminate the control connection.

For "LIST" command, the server the format the open connection records as a multi-line string, return it to the punch client through the connection, and close the connection. Each record occupies one line which is formatted by joining all fields with a space. Instead of leaving the fields about the bytes received and sent by the server as the straight integers, they are converted in human readable units, which includes "B", "KB", "MB" and "GB", according to their sizes.

To fullfill the above functionality, the punch server contains two global data structure, users and open connections. Users is an array of the same user object structure used in the above punch config and parsed from the same config file in local path `HOME/.hole_punch/users`, which represents all the allowed punch client credentials so that the server can use it to authenticate the client "OPEN" request. The open connections is used to record the status of all the existing punch client connections. It is an array of a structure called connection, which contains all the information of a punch client session. The structure connection has six fields: username, outward facing port, client facing  port, client ip, bytes received and bytes sent. Username is the account the punch client used to authenticate. Outward facing port is the port the client request to open. The client facing port is a dedicated port which the punch server randomly choose to invite the punch client connections for every incoming request. During a punch server and client session, the client facing port remains the same. Byte received is the number of bytes the punch server received from all external connnections to the outward facing port. The byte sent is the number of bytes the punch server sent to all external connnections. Each connection structure has an extra mutex to protect it from multi-thread updates, because there may exist multiple external connection to the same port concurrently. Therefore, the server must lock the mutex before updating the connection object and unlock it afterward.


## Limitation

For every initial connection to the punch server from a client, the server only reads from it once for command. The connection closes after replying "LIST" request or receiving anything other than "OPEN" command. Even the connection keeps open as the control connection for "OPEN" command, the server no longer reads from it. Therefore, the client has only one chance to send command and cannot retry. The client can neither request both "OPEN" and "LIST" nor multiple "LIST" in one continuous connection.

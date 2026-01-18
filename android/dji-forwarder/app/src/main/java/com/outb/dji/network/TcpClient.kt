package com.outb.dji.network

import android.util.Log
import com.outb.dji.model.DroneState
import com.outb.dji.model.Message
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import java.io.DataInputStream
import java.io.DataOutputStream
import java.net.InetSocketAddress
import java.net.Socket
import java.nio.ByteBuffer
import java.util.concurrent.ConcurrentLinkedQueue

/**
 * Connection state enum
 */
enum class ConnectionState {
    DISCONNECTED,
    CONNECTING,
    CONNECTED,
    RECONNECTING
}

/**
 * TCP Client for communicating with OUTB gateway
 */
class TcpClient(
    private val deviceId: String,
    private val sdkVersion: String = "5.7.1"
) {
    companion object {
        private const val TAG = "TcpClient"
        private const val CONNECT_TIMEOUT = 10000 // 10 seconds
        private const val HEARTBEAT_INTERVAL = 30000L // 30 seconds
        private const val RECONNECT_DELAY = 5000L // 5 seconds
        private const val MAX_MESSAGE_SIZE = 65536 // 64KB
    }

    private var socket: Socket? = null
    private var outputStream: DataOutputStream? = null
    private var inputStream: DataInputStream? = null

    private var serverHost: String = ""
    private var serverPort: Int = 14560

    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var heartbeatJob: Job? = null
    private var receiveJob: Job? = null
    private var reconnectJob: Job? = null

    private val messageQueue = ConcurrentLinkedQueue<Message>()
    private var sendJob: Job? = null

    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: StateFlow<ConnectionState> = _connectionState

    private val _lastError = MutableStateFlow<String?>(null)
    val lastError: StateFlow<String?> = _lastError

    private val json = Json {
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    /**
     * Connect to OUTB gateway
     */
    suspend fun connect(host: String, port: Int): Boolean {
        serverHost = host
        serverPort = port

        return withContext(Dispatchers.IO) {
            try {
                _connectionState.value = ConnectionState.CONNECTING
                Log.i(TAG, "Connecting to $host:$port")

                socket = Socket().apply {
                    connect(InetSocketAddress(host, port), CONNECT_TIMEOUT)
                    soTimeout = 60000 // 60 second read timeout
                }

                outputStream = DataOutputStream(socket!!.getOutputStream())
                inputStream = DataInputStream(socket!!.getInputStream())

                // Send hello message
                sendMessageDirect(Message.hello(deviceId, sdkVersion))

                // Wait for ACK
                val response = receiveMessage()
                if (response?.type != "ack") {
                    throw Exception("Expected ACK, got: ${response?.type}")
                }

                _connectionState.value = ConnectionState.CONNECTED
                _lastError.value = null
                Log.i(TAG, "Connected successfully")

                // Start background jobs
                startHeartbeat()
                startReceiveLoop()
                startSendLoop()

                true
            } catch (e: Exception) {
                Log.e(TAG, "Connection failed: ${e.message}")
                _lastError.value = e.message
                _connectionState.value = ConnectionState.DISCONNECTED
                closeSocket()
                false
            }
        }
    }

    /**
     * Disconnect from server
     */
    fun disconnect() {
        Log.i(TAG, "Disconnecting")
        stopAllJobs()
        closeSocket()
        _connectionState.value = ConnectionState.DISCONNECTED
    }

    /**
     * Send drone state to server
     */
    fun sendState(state: DroneState) {
        if (_connectionState.value == ConnectionState.CONNECTED) {
            messageQueue.offer(Message.state(state))
        }
    }

    /**
     * Enable auto-reconnect
     */
    fun enableAutoReconnect() {
        scope.launch {
            connectionState.collect { state ->
                if (state == ConnectionState.DISCONNECTED && serverHost.isNotEmpty()) {
                    delay(RECONNECT_DELAY)
                    if (_connectionState.value == ConnectionState.DISCONNECTED) {
                        Log.i(TAG, "Attempting reconnect...")
                        _connectionState.value = ConnectionState.RECONNECTING
                        connect(serverHost, serverPort)
                    }
                }
            }
        }
    }

    private fun startHeartbeat() {
        heartbeatJob = scope.launch {
            while (isActive && _connectionState.value == ConnectionState.CONNECTED) {
                delay(HEARTBEAT_INTERVAL)
                try {
                    sendMessageDirect(Message.heartbeat())
                    Log.d(TAG, "Heartbeat sent")
                } catch (e: Exception) {
                    Log.e(TAG, "Heartbeat failed: ${e.message}")
                    handleConnectionError()
                }
            }
        }
    }

    private fun startReceiveLoop() {
        receiveJob = scope.launch {
            while (isActive && _connectionState.value == ConnectionState.CONNECTED) {
                try {
                    val message = receiveMessage()
                    if (message != null) {
                        Log.d(TAG, "Received: ${message.type}")
                    }
                } catch (e: Exception) {
                    if (isActive) {
                        Log.e(TAG, "Receive error: ${e.message}")
                        handleConnectionError()
                    }
                    break
                }
            }
        }
    }

    private fun startSendLoop() {
        sendJob = scope.launch {
            while (isActive && _connectionState.value == ConnectionState.CONNECTED) {
                val message = messageQueue.poll()
                if (message != null) {
                    try {
                        sendMessageDirect(message)
                    } catch (e: Exception) {
                        Log.e(TAG, "Send error: ${e.message}")
                        handleConnectionError()
                        break
                    }
                } else {
                    delay(10) // Small delay when queue is empty
                }
            }
        }
    }

    private suspend fun sendMessageDirect(message: Message) {
        withContext(Dispatchers.IO) {
            val jsonStr = json.encodeToString(message)
            val bytes = jsonStr.toByteArray(Charsets.UTF_8)

            synchronized(this@TcpClient) {
                outputStream?.let { out ->
                    // Write length prefix (4 bytes, big endian)
                    out.writeInt(bytes.size)
                    // Write JSON data
                    out.write(bytes)
                    out.flush()
                }
            }
        }
    }

    private suspend fun receiveMessage(): Message? {
        return withContext(Dispatchers.IO) {
            inputStream?.let { input ->
                // Read length prefix
                val length = input.readInt()
                if (length > MAX_MESSAGE_SIZE) {
                    throw Exception("Message too large: $length")
                }

                // Read JSON data
                val bytes = ByteArray(length)
                input.readFully(bytes)
                val jsonStr = String(bytes, Charsets.UTF_8)

                json.decodeFromString<Message>(jsonStr)
            }
        }
    }

    private fun handleConnectionError() {
        if (_connectionState.value == ConnectionState.CONNECTED) {
            _connectionState.value = ConnectionState.DISCONNECTED
            closeSocket()
        }
    }

    private fun stopAllJobs() {
        heartbeatJob?.cancel()
        receiveJob?.cancel()
        sendJob?.cancel()
        reconnectJob?.cancel()
    }

    private fun closeSocket() {
        try {
            outputStream?.close()
            inputStream?.close()
            socket?.close()
        } catch (e: Exception) {
            Log.e(TAG, "Error closing socket: ${e.message}")
        }
        outputStream = null
        inputStream = null
        socket = null
    }

    /**
     * Release all resources
     */
    fun release() {
        disconnect()
        scope.cancel()
    }
}

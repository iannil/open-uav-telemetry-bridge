package com.outb.dji.service

import android.app.*
import android.content.Intent
import android.os.Binder
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import com.outb.dji.R
import com.outb.dji.dji.DJIConnectionState
import com.outb.dji.dji.DJIManager
import com.outb.dji.network.ConnectionState
import com.outb.dji.network.TcpClient
import com.outb.dji.ui.MainActivity
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine

/**
 * Foreground service for continuous telemetry forwarding
 */
class ForwarderService : Service() {

    companion object {
        private const val TAG = "ForwarderService"
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL_ID = "outb_forwarder_channel"
        private const val DEFAULT_SEND_RATE_HZ = 1.0
    }

    private val binder = LocalBinder()
    private val scope = CoroutineScope(Dispatchers.Default + SupervisorJob())

    private lateinit var djiManager: DJIManager
    private var tcpClient: TcpClient? = null

    private var forwardingJob: Job? = null
    private var sendRateHz = DEFAULT_SEND_RATE_HZ

    // Configuration
    private var serverHost = ""
    private var serverPort = 14560
    private var simulationMode = false

    private val _isForwarding = MutableStateFlow(false)
    val isForwarding: StateFlow<Boolean> = _isForwarding

    private val _statistics = MutableStateFlow(ForwardingStats())
    val statistics: StateFlow<ForwardingStats> = _statistics

    inner class LocalBinder : Binder() {
        fun getService(): ForwarderService = this@ForwarderService
    }

    override fun onCreate() {
        super.onCreate()
        Log.i(TAG, "Service created")
        createNotificationChannel()
        djiManager = DJIManager(applicationContext)
    }

    override fun onBind(intent: Intent): IBinder = binder

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        Log.i(TAG, "Service started")
        startForeground(NOTIFICATION_ID, createNotification("准备中..."))
        return START_STICKY
    }

    override fun onDestroy() {
        Log.i(TAG, "Service destroyed")
        stopForwarding()
        scope.cancel()
        super.onDestroy()
    }

    /**
     * Configure the service
     */
    fun configure(
        host: String,
        port: Int,
        rateHz: Double = DEFAULT_SEND_RATE_HZ,
        simulation: Boolean = false
    ) {
        serverHost = host
        serverPort = port
        sendRateHz = rateHz
        simulationMode = simulation
        Log.i(TAG, "Configured: $host:$port @ ${rateHz}Hz, simulation=$simulation")
    }

    /**
     * Start forwarding telemetry data
     */
    fun startForwarding() {
        if (_isForwarding.value) {
            Log.w(TAG, "Already forwarding")
            return
        }

        scope.launch {
            try {
                // Initialize DJI or simulation
                if (simulationMode) {
                    djiManager.setSimulationMode("dji")
                } else {
                    djiManager.initialize()
                    djiManager.startTelemetryCollection()
                }

                // Create and connect TCP client
                val deviceId = djiManager.getDeviceId()
                tcpClient = TcpClient(deviceId)

                updateNotification("正在连接 $serverHost...")

                val connected = tcpClient!!.connect(serverHost, serverPort)
                if (!connected) {
                    Log.e(TAG, "Failed to connect to server")
                    updateNotification("连接失败")
                    return@launch
                }

                tcpClient!!.enableAutoReconnect()
                _isForwarding.value = true

                updateNotification("正在转发数据")

                // Start forwarding loop
                startForwardingLoop()

            } catch (e: Exception) {
                Log.e(TAG, "Start forwarding error: ${e.message}")
                updateNotification("错误: ${e.message}")
            }
        }
    }

    /**
     * Stop forwarding telemetry data
     */
    fun stopForwarding() {
        _isForwarding.value = false
        forwardingJob?.cancel()
        tcpClient?.disconnect()
        tcpClient?.release()
        tcpClient = null
        djiManager.stopTelemetryCollection()
        updateNotification("已停止")
        Log.i(TAG, "Forwarding stopped")
    }

    private fun startForwardingLoop() {
        val intervalMs = (1000.0 / sendRateHz).toLong()

        forwardingJob = scope.launch {
            var messagesSent = 0L
            val startTime = System.currentTimeMillis()

            while (isActive && _isForwarding.value) {
                try {
                    // Get drone state
                    val state = if (simulationMode) {
                        djiManager.generateSimulatedState()
                    } else {
                        djiManager.getCurrentState()
                    }

                    // Send if connected and have data
                    if (state != null && tcpClient?.connectionState?.value == ConnectionState.CONNECTED) {
                        tcpClient?.sendState(state)
                        messagesSent++

                        // Update statistics
                        val elapsed = System.currentTimeMillis() - startTime
                        _statistics.value = ForwardingStats(
                            messagesSent = messagesSent,
                            elapsedMs = elapsed,
                            avgRateHz = if (elapsed > 0) messagesSent * 1000.0 / elapsed else 0.0,
                            lastState = state
                        )
                    }

                    delay(intervalMs)
                } catch (e: Exception) {
                    Log.e(TAG, "Forwarding error: ${e.message}")
                    delay(1000) // Wait before retry
                }
            }
        }
    }

    /**
     * Get DJI connection state
     */
    fun getDJIConnectionState(): StateFlow<DJIConnectionState> = djiManager.connectionState

    /**
     * Get network connection state
     */
    fun getNetworkConnectionState(): StateFlow<ConnectionState>? = tcpClient?.connectionState

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "OUTB Forwarder",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "DJI 遥测数据转发服务"
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(status: String): Notification {
        val intent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("OUTB DJI Forwarder")
            .setContentText(status)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(status: String) {
        val notification = createNotification(status)
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification)
    }
}

/**
 * Statistics for forwarding session
 */
data class ForwardingStats(
    val messagesSent: Long = 0,
    val elapsedMs: Long = 0,
    val avgRateHz: Double = 0.0,
    val lastState: com.outb.dji.model.DroneState? = null
)

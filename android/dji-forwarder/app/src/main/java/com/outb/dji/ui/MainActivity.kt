package com.outb.dji.ui

import android.Manifest
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.os.IBinder
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.lifecycleScope
import com.outb.dji.databinding.ActivityMainBinding
import com.outb.dji.dji.DJIConnectionState
import com.outb.dji.network.ConnectionState
import com.outb.dji.service.ForwarderService
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {

    companion object {
        private const val PERMISSION_REQUEST_CODE = 1001
    }

    private lateinit var binding: ActivityMainBinding
    private var forwarderService: ForwarderService? = null
    private var serviceBound = false

    private val serviceConnection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, binder: IBinder?) {
            val localBinder = binder as ForwarderService.LocalBinder
            forwarderService = localBinder.getService()
            serviceBound = true
            observeServiceState()
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            forwarderService = null
            serviceBound = false
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupUI()
        checkPermissions()
    }

    override fun onStart() {
        super.onStart()
        // Bind to service
        Intent(this, ForwarderService::class.java).also { intent ->
            bindService(intent, serviceConnection, Context.BIND_AUTO_CREATE)
        }
    }

    override fun onStop() {
        super.onStop()
        if (serviceBound) {
            unbindService(serviceConnection)
            serviceBound = false
        }
    }

    private fun setupUI() {
        // Set default values
        binding.editServerHost.setText("192.168.1.100")
        binding.editServerPort.setText("14560")
        binding.editSendRate.setText("1.0")
        binding.switchSimulation.isChecked = true // Default to simulation for testing

        // Connect button
        binding.btnConnect.setOnClickListener {
            if (forwarderService?.isForwarding?.value == true) {
                stopForwarding()
            } else {
                startForwarding()
            }
        }

        // Update UI based on state
        updateUI(false)
    }

    private fun startForwarding() {
        val host = binding.editServerHost.text.toString().trim()
        val port = binding.editServerPort.text.toString().toIntOrNull() ?: 14560
        val rate = binding.editSendRate.text.toString().toDoubleOrNull() ?: 1.0
        val simulation = binding.switchSimulation.isChecked

        if (host.isEmpty()) {
            Toast.makeText(this, "请输入服务器地址", Toast.LENGTH_SHORT).show()
            return
        }

        // Start foreground service
        val serviceIntent = Intent(this, ForwarderService::class.java)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(serviceIntent)
        } else {
            startService(serviceIntent)
        }

        // Configure and start forwarding
        forwarderService?.configure(host, port, rate, simulation)
        forwarderService?.startForwarding()
    }

    private fun stopForwarding() {
        forwarderService?.stopForwarding()
    }

    private fun observeServiceState() {
        lifecycleScope.launch {
            forwarderService?.isForwarding?.collectLatest { isForwarding ->
                updateUI(isForwarding)
            }
        }

        lifecycleScope.launch {
            forwarderService?.statistics?.collectLatest { stats ->
                binding.tvMessagesSent.text = "已发送: ${stats.messagesSent}"
                binding.tvAvgRate.text = "频率: %.2f Hz".format(stats.avgRateHz)

                stats.lastState?.let { state ->
                    binding.tvLocation.text = "位置: %.6f, %.6f".format(
                        state.location.lat, state.location.lon
                    )
                    binding.tvAltitude.text = "高度: %.1f m".format(state.location.altGnss)
                    binding.tvBattery.text = "电量: ${state.status.batteryPercent}%"
                    binding.tvFlightMode.text = "模式: ${state.status.flightMode}"
                }
            }
        }

        lifecycleScope.launch {
            forwarderService?.getDJIConnectionState()?.collectLatest { state ->
                binding.tvDjiStatus.text = when (state) {
                    DJIConnectionState.DISCONNECTED -> "DJI: 未连接"
                    DJIConnectionState.CONNECTING -> "DJI: 连接中..."
                    DJIConnectionState.CONNECTED -> "DJI: 已连接"
                    DJIConnectionState.PRODUCT_CONNECTED -> "DJI: 飞机已连接"
                }
            }
        }

        lifecycleScope.launch {
            forwarderService?.getNetworkConnectionState()?.collectLatest { state ->
                binding.tvNetworkStatus.text = when (state) {
                    ConnectionState.DISCONNECTED -> "网络: 未连接"
                    ConnectionState.CONNECTING -> "网络: 连接中..."
                    ConnectionState.CONNECTED -> "网络: 已连接"
                    ConnectionState.RECONNECTING -> "网络: 重连中..."
                    null -> "网络: -"
                }
            }
        }
    }

    private fun updateUI(isForwarding: Boolean) {
        binding.btnConnect.text = if (isForwarding) "停止" else "开始"
        binding.editServerHost.isEnabled = !isForwarding
        binding.editServerPort.isEnabled = !isForwarding
        binding.editSendRate.isEnabled = !isForwarding
        binding.switchSimulation.isEnabled = !isForwarding
    }

    private fun checkPermissions() {
        val permissions = mutableListOf(
            Manifest.permission.INTERNET,
            Manifest.permission.ACCESS_NETWORK_STATE,
            Manifest.permission.ACCESS_WIFI_STATE
        )

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissions.add(Manifest.permission.POST_NOTIFICATIONS)
        }

        val notGranted = permissions.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }

        if (notGranted.isNotEmpty()) {
            ActivityCompat.requestPermissions(
                this,
                notGranted.toTypedArray(),
                PERMISSION_REQUEST_CODE
            )
        }
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == PERMISSION_REQUEST_CODE) {
            val denied = grantResults.count { it != PackageManager.PERMISSION_GRANTED }
            if (denied > 0) {
                Toast.makeText(this, "部分权限被拒绝，功能可能受限", Toast.LENGTH_LONG).show()
            }
        }
    }
}

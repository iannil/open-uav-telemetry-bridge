package com.outb.dji.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Unified drone state model matching OUTB protocol
 */
@Serializable
data class DroneState(
    @SerialName("device_id")
    val deviceId: String,

    @SerialName("timestamp")
    val timestamp: Long,

    @SerialName("protocol_source")
    val protocolSource: String = "dji",

    @SerialName("location")
    val location: Location,

    @SerialName("attitude")
    val attitude: Attitude,

    @SerialName("status")
    val status: Status,

    @SerialName("velocity")
    val velocity: Velocity
)

@Serializable
data class Location(
    @SerialName("lat")
    val lat: Double,

    @SerialName("lon")
    val lon: Double,

    @SerialName("alt_baro")
    val altBaro: Double,

    @SerialName("alt_gnss")
    val altGnss: Double,

    @SerialName("coordinate_system")
    val coordinateSystem: String = "WGS84"
)

@Serializable
data class Attitude(
    @SerialName("roll")
    val roll: Double,

    @SerialName("pitch")
    val pitch: Double,

    @SerialName("yaw")
    val yaw: Double
)

@Serializable
data class Status(
    @SerialName("battery_percent")
    val batteryPercent: Int,

    @SerialName("flight_mode")
    val flightMode: String,

    @SerialName("armed")
    val armed: Boolean,

    @SerialName("signal_quality")
    val signalQuality: Int
)

@Serializable
data class Velocity(
    @SerialName("vx")
    val vx: Double,

    @SerialName("vy")
    val vy: Double,

    @SerialName("vz")
    val vz: Double
)

/**
 * Factory function to create empty DroneState
 */
fun emptyDroneState(deviceId: String): DroneState = DroneState(
    deviceId = deviceId,
    timestamp = System.currentTimeMillis(),
    location = Location(0.0, 0.0, 0.0, 0.0),
    attitude = Attitude(0.0, 0.0, 0.0),
    status = Status(0, "UNKNOWN", false, 0),
    velocity = Velocity(0.0, 0.0, 0.0)
)

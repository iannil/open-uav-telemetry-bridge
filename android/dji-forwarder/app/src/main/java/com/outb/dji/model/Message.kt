package com.outb.dji.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement

/**
 * Message types for OUTB protocol
 */
enum class MessageType {
    HELLO,
    STATE,
    HEARTBEAT,
    ACK
}

/**
 * Base message for OUTB communication
 */
@Serializable
data class Message(
    @SerialName("type")
    val type: String,

    @SerialName("device_id")
    val deviceId: String? = null,

    @SerialName("sdk_version")
    val sdkVersion: String? = null,

    @SerialName("timestamp")
    val timestamp: Long? = null,

    @SerialName("data")
    val data: JsonElement? = null
) {
    companion object {
        fun hello(deviceId: String, sdkVersion: String): Message = Message(
            type = "hello",
            deviceId = deviceId,
            sdkVersion = sdkVersion
        )

        fun state(droneState: DroneState): Message = Message(
            type = "state",
            data = kotlinx.serialization.json.Json.encodeToJsonElement(
                DroneState.serializer(),
                droneState
            )
        )

        fun heartbeat(): Message = Message(
            type = "heartbeat",
            timestamp = System.currentTimeMillis()
        )
    }
}

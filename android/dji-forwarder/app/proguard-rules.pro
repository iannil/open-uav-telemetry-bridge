# Add project specific ProGuard rules here.

# Keep Kotlin serialization
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.AnnotationsKt

-keepclassmembers class kotlinx.serialization.json.** {
    *** Companion;
}
-keepclasseswithmembers class kotlinx.serialization.json.** {
    kotlinx.serialization.KSerializer serializer(...);
}

-keep,includedescriptorclasses class com.outb.dji.model.**$$serializer { *; }
-keepclassmembers class com.outb.dji.model.** {
    *** Companion;
}
-keepclasseswithmembers class com.outb.dji.model.** {
    kotlinx.serialization.KSerializer serializer(...);
}

# DJI SDK rules (add when SDK is integrated)
# -keep class dji.** { *; }
# -dontwarn dji.**

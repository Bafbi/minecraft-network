plugins {
    id("java")
    id("com.gradleup.shadow") version "8.3.0"
}

group = "fr.bafbi"
version = "0.1.0"

repositories {
    mavenCentral()
}

dependencies {
    // Change this to the latest version
    implementation("net.minestom:minestom-snapshots:1_21_4-7599413490")
    implementation("ch.qos.logback:logback-classic:1.5.6")
    implementation("io.nats:jnats:2.21.1")
    implementation("io.vavr:vavr:0.10.5")
    implementation("com.google.code.gson:gson:2.13.0")
}

java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(21)) // Minestom has a minimum Java version of 21
    }
}

tasks {
    jar {
        manifest {
            attributes["Main-Class"] = "fr.bafbi.lobby.Main" // Change this to your main class
        }
    }

    build {
        dependsOn(shadowJar)
    }
    shadowJar {
        mergeServiceFiles()
        archiveClassifier.set("") // Prevent the -all suffix on the shadowjar file.
    }
}

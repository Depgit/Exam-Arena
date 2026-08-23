package com.examarena

import android.app.Application
import com.examarena.core.di.AppContainer

class ExamArenaApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
    }
}

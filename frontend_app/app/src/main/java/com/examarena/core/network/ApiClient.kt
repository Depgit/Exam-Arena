package com.examarena.core.network

import com.examarena.core.common.Resource
import com.examarena.core.datastore.SessionManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import java.io.IOException
import java.util.concurrent.TimeUnit

class ApiClient(
    private val sessionManager: SessionManager,
    private val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(20, TimeUnit.SECONDS)
        .writeTimeout(20, TimeUnit.SECONDS)
        .build()
) {
    val json = Json {
        ignoreUnknownKeys = true
        isLenient = true
        encodeDefaults = true
        prettyPrint = false
    }

    private val jsonMediaType = "application/json; charset=utf-8".toMediaType()

    suspend fun <T> get(
        path: String,
        requireAuth: Boolean = true,
        deserializer: (String) -> T
    ): Resource<T> = withContext(Dispatchers.IO) {
        val baseUrl = sessionManager.getApiUrl().trimEnd('/')
        val url = if (path.startsWith("http")) path else "$baseUrl$path"

        val requestBuilder = Request.Builder().url(url).get()

        if (requireAuth) {
            val token = sessionManager.getAuthToken()
            if (token != null) {
                requestBuilder.addHeader("Authorization", "Bearer $token")
            }
        }

        executeRequest(requestBuilder.build(), deserializer)
    }

    suspend fun <T> post(
        path: String,
        bodyJson: String = "{}",
        requireAuth: Boolean = true,
        deserializer: (String) -> T
    ): Resource<T> = withContext(Dispatchers.IO) {
        val baseUrl = sessionManager.getApiUrl().trimEnd('/')
        val url = if (path.startsWith("http")) path else "$baseUrl$path"

        val requestBody = bodyJson.toRequestBody(jsonMediaType)
        val requestBuilder = Request.Builder().url(url).post(requestBody)

        if (requireAuth) {
            val token = sessionManager.getAuthToken()
            if (token != null) {
                requestBuilder.addHeader("Authorization", "Bearer $token")
            }
        }

        executeRequest(requestBuilder.build(), deserializer)
    }

    suspend fun <T> delete(
        path: String,
        requireAuth: Boolean = true,
        deserializer: (String) -> T
    ): Resource<T> = withContext(Dispatchers.IO) {
        val baseUrl = sessionManager.getApiUrl().trimEnd('/')
        val url = if (path.startsWith("http")) path else "$baseUrl$path"

        val requestBuilder = Request.Builder().url(url).delete()

        if (requireAuth) {
            val token = sessionManager.getAuthToken()
            if (token != null) {
                requestBuilder.addHeader("Authorization", "Bearer $token")
            }
        }

        executeRequest(requestBuilder.build(), deserializer)
    }

    private fun <T> executeRequest(
        request: Request,
        deserializer: (String) -> T
    ): Resource<T> {
        return try {
            val response: Response = okHttpClient.newCall(request).execute()
            val responseBody = response.body?.string() ?: ""

            if (response.isSuccessful) {
                val data = deserializer(responseBody)
                Resource.Success(data)
            } else {
                val errorMsg = parseErrorMessage(responseBody, response.code)
                Resource.Error(errorMsg)
            }
        } catch (e: IOException) {
            Resource.Error("Network error. Please check server connection: ${e.localizedMessage ?: "Unknown error"}", e)
        } catch (e: Exception) {
            Resource.Error("An unexpected error occurred: ${e.localizedMessage ?: "Unknown error"}", e)
        }
    }

    private fun parseErrorMessage(responseBody: String, statusCode: Int): String {
        return try {
            val jsonObject = json.parseToJsonElement(responseBody).jsonObject
            jsonObject["error"]?.jsonPrimitive?.content
                ?: jsonObject["message"]?.jsonPrimitive?.content
                ?: "Request failed with status $statusCode"
        } catch (e: Exception) {
            if (responseBody.isNotBlank()) responseBody else "Request failed with status $statusCode"
        }
    }
}

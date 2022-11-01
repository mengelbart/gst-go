#include "gst.h"
#include <gst/app/gstappsrc.h>

GMainLoop* create_mainloop() {
    GMainLoop* main_loop = NULL;
    return g_main_loop_new(NULL, FALSE);
}

void start_mainloop(GMainLoop* main_loop) {
    g_main_loop_run(main_loop);
}

void stop_mainloop(GMainLoop* main_loop) {
    g_main_loop_quit(main_loop);
}

GstFlowReturn new_sample_handler(GstElement* object, gpointer user_data) {
    GstSample* sample = NULL;
    GstBuffer* buffer = NULL;
    gpointer copy = NULL;
    gsize copy_size = 0;
    SampleHandlerUserData* s = (SampleHandlerUserData*) user_data;

    g_signal_emit_by_name (object, "pull-sample", &sample);

    if (sample) {
        buffer = gst_sample_get_buffer(sample);
        if (buffer) {
            int duration = -1;
            if (buffer->duration != GST_CLOCK_TIME_NONE) {
                duration = (int) buffer->duration;
            }
            gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);
            goHandlePipelineBuffer(copy, copy_size, duration, s->pipelineId);
        }
        gst_sample_unref(sample);
    }
    return GST_FLOW_OK;
}

static gboolean bus_call(GstBus* bus, GstMessage* msg, gpointer user_data) {
    SampleHandlerUserData* s = (SampleHandlerUserData*) user_data;

    switch (GST_MESSAGE_TYPE(msg)) {
    case GST_MESSAGE_EOS: {
        goHandleBusCall(s->pipelineId, 0, NULL);
        break;
    }

    case GST_MESSAGE_ERROR: {
        gchar* debug;
        GError* error;

        gst_message_parse_error(msg, &error, &debug);
        goHandleBusCall(s->pipelineId, 1, error->message);
        g_free(debug);
        g_error_free(error);
        break;
    }

    default:
        break;
    }
    return TRUE;
}

GstElement* create_pipeline(char* pipelineStr) {
    GError* error = NULL;
    gst_init(NULL, NULL);
    return gst_parse_launch(pipelineStr, &error);
}

void start_pipeline(GstElement* pipeline, int pipelineId) {
    SampleHandlerUserData* s = malloc(sizeof(SampleHandlerUserData));
    s->pipelineId = pipelineId;

    GstBus* bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
    gst_bus_add_watch(bus, bus_call, s);
    gst_object_unref(bus);

    gst_element_set_state(pipeline, GST_STATE_PLAYING);
}

void link_appsink(GstElement* pipeline, int pipelineId) {
    SampleHandlerUserData* s = malloc(sizeof(SampleHandlerUserData));
    s->pipelineId = pipelineId;

    GstElement* appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
    g_object_set(appsink, "emit-signals", TRUE, NULL);
    g_signal_connect(appsink, "new-sample", G_CALLBACK(new_sample_handler), s);
    gst_object_unref(appsink);
}

void push_buffer(GstElement* pipeline, void* buffer, int len) {
    GstElement* src = gst_bin_get_by_name(GST_BIN(pipeline), "src");
    if (src != NULL) {
        gpointer p = g_memdup(buffer, len);
        GstBuffer* buffer = gst_buffer_new_wrapped(p, len);
        gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
        gst_object_unref(src);
    }
}

void stop_pipeline(GstElement* pipeline) {
    gst_element_send_event(pipeline, gst_event_new_eos());
}

void destroy_pipeline(GstElement* pipeline) {
    gst_element_set_state(pipeline, GST_STATE_NULL);
    gst_object_unref(pipeline);
}

unsigned int get_property_uint(GstElement* pipeline, char* name, char* prop) {
    GstElement* element;
    element = gst_bin_get_by_name(GST_BIN(pipeline), name);
    unsigned int value = 0;

    if (element) {
        g_object_get(element, prop, &value, NULL);
    }
    return value;
}

void set_property_uint(GstElement* pipeline, char* name, char* prop, unsigned int value) {
    GstElement* element;
    element = gst_bin_get_by_name(GST_BIN(pipeline), name);

    if (element) {
        g_object_set(element, prop, value, NULL);
        gst_object_unref(element);
    }
}

void dump_pipeline_graph(GstElement* pipeline, char* filename) {
    GST_DEBUG_BIN_TO_DOT_FILE(GST_BIN(pipeline), GST_DEBUG_GRAPH_SHOW_ALL, filename);
}

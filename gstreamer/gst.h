#ifndef GST_H
#define GST_H

#include <gst/gst.h>

typedef struct SampleHandlerUserData {
    int pipelineId;
} SampleHandlerUserData;

extern void goHandlePipelineBuffer(void* buffer, int bufferLen, int pipelineId);
extern void goHandleBusCall(int pipelineId, int signal, char* message);

GMainLoop* create_mainloop();
void start_mainloop(GMainLoop* main_loop);
void stop_mainloop(GMainLoop* main_loop);

void dump_pipeline_graph(GstElement* pipeline, char* filename);

GstElement* create_pipeline(char* pipelineStr);
void start_pipeline(GstElement* pipeline, int pipelineId);
void stop_pipeline(GstElement* pipeline);
void destroy_pipeline(GstElement* pipeline);

void link_appsink(GstElement* pipeline, int pipelineId);
void push_buffer(GstElement* pipeline, void* buffer, int len);

unsigned int get_property_uint(GstElement* pipeline, char* name, char* prop);
void set_property_uint(GstElement* pipeline, char* name, char* prop, unsigned int value);


#endif /* #ifndef GST_H */
